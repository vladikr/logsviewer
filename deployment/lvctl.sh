#!/bin/bash


Help()
{
   # Usage
   echo "Control an instance of the logsViewer"
   echo
   echo "Syntax: lvctl [command] [options]"
   echo "commands:"
   echo "c|create     Create a new instance"
   echo "d|delete     Delete instace"
   echo "h|help       Print this Help."
   echo
   echo "options:"
   echo "i|suffix            Instance ID"
   echo "s|storage_class     Storage Class to use"
   echo "t|logsviewer_image  LogsViewer image to use"
   echo
}

function get_pod_status() {
  local route="${1}" ; shift
  oc get pod ${route} -o jsonpath={.status.phase}
}

function wait_for_route() {
  local route="${1}" ; shift
  local timeout="${1}" ; shift
  sleep 3
    SECONDS=0
    echo -n "Waiting for ${route} route ..."
    OCP4_REGISTER=$(oc get route ${route} -o jsonpath={.status.ingress[0].host})
    while ! [ -z "${OCP4_REGISTER}"  ] ; do
      echo -n "."
      if [ "${SECONDS}" -gt "${timeout}0" ]; then
        echo " FAIL"
        return 1
      fi
      sleep 3
    done
    echo "${route} - ${OCP4_REGISTER}"
}

function wait_for_pod() {
  local route="${1}" ; shift
  local timeout="${1}" ; shift
  sleep 3
    SECONDS=0
    echo -n "Waiting for ${route} pod ..."
    while ! [ "$(get_pod_status "${route}")" == "Running"  ] ; do
      echo -n "."
      if [ "${SECONDS}" -gt "${timeout}0" ]; then
        echo " FAIL"
        return 1
      fi
      sleep 3
    done
    echo "DONE ${route}"
}


function create_instance() {
    local ID=$(cat /dev/urandom | tr -dc 'a-za-z0-9' | fold -w 6 | head -n 1)
    if [[ $storage_class != "" ]]; then
        add_st_class="-p STORAGE_CLASS=${storage_class}"
     fi
    if [[ $logsviewer_image != "" ]]; then
        add_tag="-p LOGSVIEWER_IMAGE=${logsviewer_image}"
     fi
    oc process -f deployment/elk_pod_template.yaml -p SUFFIX=${ID} ${add_st_class} ${add_tag}| oc create -f -
    sleep 5
    wait_for_pod "logsviewer-${ID}" 300
    oc get routes
}

function delete_instance() {
    local route="${1}" ; shift
    if [[ $storage_class != "" ]]; then
         add_st_class="-p STORAGE_CLASS=${storage_class}"
    fi
    if [[ $logsviewer_image != "" ]]; then
        add_tag="-p LOGSVIEWER_IMAGE=${logsviewer_image}"
     fi
    oc process -f deployment/elk_pod_template.yaml -p SUFFIX=${route} ${add_st_class} ${add_tag}| oc delete -f -
}

# Set variables
create=false
delete=false
storage_class=
logsviewer_image=
suffix=

TEMP=$(getopt -o cdi:s:t:h --long create,delete,suffix:,storage_class:,logsviewer_image:,help -n 'lvctl' -- "$@")
eval set -- "$TEMP"

while true; do
  case "$1" in
        -c | --create )           create=true; shift ;;
        -d | --delete )           delete=true; shift ;;
        -i | --suffix )           suffix="$2"; shift 2 ;;
        -s | --storage_class )    storage_class="$2"; shift 2 ;;
        -t | --logsviewer_image ) logsviewer_image="$2"; shift 2 ;;
        -h | --help)              Help shift; break;;
     \?) # Invalid option
         Help
         exit;;
    -- ) shift; break ;;
    * ) break ;;
   esac
done

if $create; then 
    create_instance ${storage_class}
elif $delete; then 
    delete_instance ${suffix} ${storage_class}
fi
