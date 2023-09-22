package backend

import (
    "fmt"
    "time"
    "logsviewer/pkg/backend/db"
)

func handlePVCs(res db.QueryResults) string {
    // handle PVCs
    pvcList := ""
    for _, val := range res.PVCs {
        pvcList += fmt.Sprintf(` or "%s"`, val)
    }
    return pvcList
}

func formatVMIMigrationDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:'%s'))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api" or podName: "%s" or podName: "%s" or podName: "%s" or podName: "%s" or "%s" or "%s" or "%s" or "%s"%s'),sort:!(!('@timestamp',asc)))`

    startTimeStamp := res.StartTimestamp.Add(-time.Second * 30)
    timeStamp := fmt.Sprintf("%sZ", startTimeStamp.UTC().Format("2006-01-02T15:04:05.000"))
    endTimeStamp := fmt.Sprintf("%sZ", res.EndTimestamp.UTC().Format("2006-01-02T15:04:05.000"))

    pvcsList := handlePVCs(res)
    
    migrationLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, endTimeStamp, res.SourcePod, res.SourceHandler, res.TargetPod, res.TargetHandler, res.SourcePodUUID, res.VMIUUID, res.TargetPodUUID, res.MigrationUUID, pvcsList)

    return migrationLogsQuery
}

func formatSingleVMIDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:now))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api" or podName: "%s" or podName: "%s" or "%s" or "%s"%s'),sort:!(!('@timestamp',asc)))`


    startTimeStamp := res.StartTimestamp.Add(-time.Second * 30)
    timeStamp := fmt.Sprintf("%sZ", startTimeStamp.UTC().Format("2006-01-02T15:04:05.000"))
    pvcsList := handlePVCs(res)
    
    vmiLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, res.SourcePod, res.SourceHandler, res.SourcePodUUID, res.VMIUUID, pvcsList)

    return vmiLogsQuery
}

func formatFullVMIHistoryDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:now))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api"%s%s%s or %s'),sort:!(!('@timestamp',asc)))`


    startTimeStamp := res.StartTimestamp.Add(-time.Second * 30)
    timeStamp := fmt.Sprintf("%sZ", startTimeStamp.UTC().Format("2006-01-02T15:04:05.000"))
    podNames := ""
    podUUIDs := ""

    // handle involved virt-launchers
    for _, pod := range res.InvolvedVirtLaunchers {
        podNames += fmt.Sprintf(` or podName: "%s"`, pod.Name)
        podUUIDs += fmt.Sprintf(` or "%s"`, pod.UUID)
    }

    // handle involved virt-handlers
    for _, pod := range res.InvolvedVirtHandlers {
        podNames += fmt.Sprintf(` or podName: "%s"`, pod.Name)
        podUUIDs += fmt.Sprintf(` or "%s"`, pod.UUID)
    }

    // handle PVCs
    pvcsList := handlePVCs(res)
    
    vmiLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, podNames, podUUIDs, pvcsList, res.VMIUUID)

    return vmiLogsQuery
}


func formatSinglePodDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:now))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'podName: "%s" or "%s"%s'),sort:!(!('@timestamp',asc)))`


    startTimeStamp := res.StartTimestamp.Add(-time.Second * 30)
    timeStamp := fmt.Sprintf("%sZ", startTimeStamp.UTC().Format("2006-01-02T15:04:05.000"))
    pvcsList := handlePVCs(res)
    
    podLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, res.SourcePod, res.SourcePodUUID, pvcsList)

    return podLogsQuery
}

