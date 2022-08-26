# Simple bash shell to launch WoST services

# Common configuration with dapr
# dapr configuration: config/dapr.config (option -c)
# sidecar-service communication using unix domain sockets (option -u)
# state store: mongodb
# pubsub: memory   (inter-service notification pubsub)
CONFIG="../config"

if [ "HUB_TEST" != "" ]; then
# test settings
BIN_FOLDER="../dist/bin"
DATA_FOLDER="/tmp/wost"
SOCK_FOLDER="/tmp/wost"
LOG_LEVEL="info"
DAPR_CONFIG="./config/dapr/config.yaml"
DAPR_COMPONENTS="./config/dapr/components"
else
# production settings using ~/.dapr settings folder
DATA_FOLDER="./data"
BIN_FOLDER="./bin"
SOCK_FOLDER="/run/wost"
LOG_LEVEL="warn"
DAPR_CONFIG="./config/dapr/config.yaml"
DAPR_COMPONENTS="./config/dapr/components"
fi
DAPR_SETTINGS="--log-level=$LOG_LEVEL -c $DAPR_CONFIG -d $DAPR_COMPONENTS -u $SOCK_FOLDER"

# Prepare runtime environment. This will clear all domain sockets.
mkdir -p $DATA_FOLDER
rm $SOCK_FOLDER/*.socket

# core services
dapr run  $DAPR_SETTINGS   --app-id certsvc  --app-protocol grpc --app-port 8880 \
 -- $BIN_FOLDER/certsvc -p 8880

#dapr run  $DAPR_SETTINGS   --app-id thingstore  --app-protocol grpc --app-port 8881 \
# -- $BIN_FOLDER/thingstore -a :8881 -f $DATA_FOLDER/thingstore.json &

#dapr run  $DAPR_SETTINGS   --app-id historystore  --app-protocol grpc --app-port 8882 \
# -- $BIN_FOLDER/historystore -a :8882 &



# gateways
