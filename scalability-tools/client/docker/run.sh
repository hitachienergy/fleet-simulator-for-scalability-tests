#/bin/bash

if [ $NEED_COMPILE ]; then
    echo "############## Build Plugin ##############"

    cd $PLUGIN_DIR

    cert_location=/usr/local/share/ca-certificates
    apt-get update && apt-get install -y ca-certificates openssl
    openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
    openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
    update-ca-certificates

    go build -buildmode=plugin -o ${PLUGIN_PATH} .

    echo ""
fi

cd /app

echo "############## Simulation setup ##############"

echo "OPATH=$OPATH"
echo "CLIENT_NUM=$CLIENT_NUM"
echo "IDX_OFFSET=$IDX_OFFSET"
echo "STATUS_SERVER_PORT=$STATUS_SERVER_PORT"

echo "############## Start devices simulator ##############"
./simulator --config "${CONFIG}"  --template ${PLUGIN_PATH} --opath ${OPATH} \
            --num ${CLIENT_NUM} --offset ${IDX_OFFSET} --serverport ${STATUS_SERVER_PORT} \
            --influence "${INFLUENCE}" 
            