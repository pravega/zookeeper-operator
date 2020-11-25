sudo PROJECT_NAME=zookeeper-operator REPO_PATH=github.com/pravega/$PROJECT_NAME VERSION=0.0.0-localdev GIT_SHA=0000000  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ../../golang/go/bin/go  build -o ./${PROJECT_NAME}     -ldflags "-X ${REPO_PATH}/pkg/version.Version=${VERSION} -X ${REPO_PATH}/pkg/version.GitSHA=${GIT_SHA}"     /home/caoguanglei/src/zk-operator/cmd/manager

mv manager zookeeper-operator

make server

rm -rf zookeeper-operator
