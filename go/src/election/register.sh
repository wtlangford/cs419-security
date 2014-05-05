user=$1
secret=$2

curl --cacert cacert.crt -F "name=$user" -F "secret=$2" https://cla.wlangford.net:1444/register 2>/dev/null | awk -F'"' '{print $4}'
