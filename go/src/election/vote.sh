id=$1
vn=$2
vote=$3

curl -k -F "id=$id" -F "vn=$2" -F "vote=$3" https://localhost:4000/vote
