package cleanup

const DeletionConditionLabel = "logsviewer.openshift.io/deletion-condition"
const DeletionDelayLabel = "logsviewer.openshift.io/deletion-delay"

type DeletionCondition string

const Creation DeletionCondition = "creation"
const LastMustGatherUpload DeletionCondition = "last-must-gather-upload"
const Never DeletionCondition = "never"
