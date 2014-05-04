user=$1
secret=$2

curl -k -F "name=$user" -F "secret=$2" https://localhost:1444/register 2>/dev/null | awk -F'"' '{print $4}'
