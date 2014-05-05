id=$1
vn=$2
vote=$3

curl --cacert cacert.crt -F "id=$id" -F "vn=$2" -F "vote=$3" https://ctf.wlangford.net:4000/vote
