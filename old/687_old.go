// across all accessible namespaces in the cluster. All ServiceEntries are needed because
// Istio does not distinguish where a ServiceEntry is created when routing traffic (i.e.
// a ServiceEntry can be in any namespace and it will still work).
func (a ServiceEntryAppender) getServiceEntry(service string, globalInfo *GlobalInfo) (string, bool) {
	if globalInfo.ServiceEntries == nil {
		globalInfo.ServiceEntries = make(map[string]string)

		for ns := range a.AccessibleNamespaces {
			istioCfg, err := globalInfo.Business.IstioConfig.GetIstioConfigList(business.IstioConfigCriteria{
				IncludeServiceEntries: true,
				Namespace:             ns,
			})
			checkError(err)

			for _, entry := range istioCfg.ServiceEntries {
				if entry.Spec.Hosts != nil && (entry.Spec.Location == "MESH_EXTERNAL" || entry.Spec.Location == "MESH_INTERNAL") {
					for _, host := range entry.Spec.Hosts.([]interface{}) {
						globalInfo.ServiceEntries[host.(string)] = entry.Spec.Location.(string)
					}
				}
			}
		}
		log.Tracef("Found [%v] service entries", len(globalInfo.ServiceEntries))
	}

	location, ok := globalInfo.ServiceEntries[service]
	return location, ok
}
