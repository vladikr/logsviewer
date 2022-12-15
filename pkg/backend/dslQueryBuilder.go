package backend

import (
    "fmt"

    "logsviewer/pkg/backend/db"
)


func formatVMIMigrationDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:'%s'))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api" or podName: "%s" or podName: "%s" or podName: "%s" or podName: "%s" or "%s" or "%s" or "%s" or "%s"'),sort:!(!('@timestamp',asc)))`


    timeStamp := fmt.Sprintf("%sZ", res.StartTimestamp.UTC().Format("2006-01-02T15:04:05.000"))
    endTimeStamp := fmt.Sprintf("%sZ", res.EndTimestamp.UTC().Format("2006-01-02T15:04:05.000"))
    
    migrationLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, endTimeStamp, res.SourcePod, res.SourceHandler, res.TargetPod, res.TargetHandler, res.SourcePodUUID, res.VMIUUID, res.TargetPodUUID, res.MigrationUUID)

    return migrationLogsQuery
}

func formatSingleVMIDSLQuery(res db.QueryResults) string {

    queryTemplate := `_g=(filters:!(),refreshInterval:(pause:!t,value:0),time:(from:'%s',to:now))&_a=(columns:!(msg,podName,component,uid,subcomponent,reason,enrichment_data.pod.uid,enrichment_data.host.name,level),filters:!(('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!f,type:exists,value:exists),query:(exists:(field:msg))),('$state':(store:appState),meta:(alias:!n,disabled:!f,key:msg,negate:!t,params:(query:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.'),type:phrase),query:(match_phrase:(msg:'certificate with common name !'kubevirt.io:system:client:virt-handler!' retrieved.')))),interval:auto,query:(language:kuery,query:'containerName: "virt-controller" or containerName: "virt-api" or podName: "%s" or podName: "%s" or "%s" or "%s"'),sort:!(!('@timestamp',asc)))`


    timeStamp := fmt.Sprintf("%sZ", res.StartTimestamp.UTC().Format("2006-01-02T15:04:05.000"))
    
    vmiLogsQuery := fmt.Sprintf(queryTemplate, timeStamp, res.SourcePod, res.SourceHandler, res.SourcePodUUID, res.VMIUUID)

    return vmiLogsQuery
}

