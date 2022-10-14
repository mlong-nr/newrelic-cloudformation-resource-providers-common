package nerdgraph

import (
   "fmt"
   "github.com/newrelic-experimental/newrelic-cloudformation-resource-providers-common/cferror"
   "github.com/newrelic-experimental/newrelic-cloudformation-resource-providers-common/model"
   log "github.com/sirupsen/logrus"
)

type readResponse struct {
   Data readData `json:"data"`
}
type readData struct {
   Actor readActor `json:"actor"`
}
type readActor struct {
   Entity *readEntity `json:"entity,omitempty"`
}
type readEntity struct {
   Guid string `json:"guid"`
   Name string `json:"name"`
}

func (i *nerdgraph) Read(m model.Model) (err error) {
   defer func() {
      log.Debugf("Read: returning value: %+v type: %T", err, err)
   }()

   variables := m.GetVariables()
   i.config.InjectIntoMap(&variables)
   mutation := m.GetReadQuery()

   // Render the mutation
   mutation, err = model.Render(mutation, variables)
   if err != nil {
      log.Errorf("Read: %v", err)
      return fmt.Errorf("%w %s", &cferror.InvalidRequest{}, err.Error())
   }
   log.Debugln("Read: rendered mutation: ", mutation)
   log.Debugln("")

   // Validate mutation
   err = model.Validate(&mutation)
   if err != nil {
      log.Errorf("Read: %v", err)
      return fmt.Errorf("%w %s", &cferror.InvalidRequest{}, err.Error())
   }

   body, err := i.emit(mutation, *i.config.APIKey, i.config.GetEndpoint())
   if err != nil {
      return
   }

   key := m.GetResultKey(model.Read)
   if key != "" {
      var v interface{}
      v, err = findKeyValue(body, key)
      if err != nil {
         log.Errorf("error finding result key: %s in response: %s", key, string(body))
         return
      }

      if v == nil {
         log.Errorf("Read: result not returned by NerdGraph operation")
         err = fmt.Errorf("%w Read: result not returned by NerdGraph operation", &cferror.InvalidRequest{})
         return
      }
   }
   return
}