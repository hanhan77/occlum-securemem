FROM occlum/occlum:0.31.0-ubuntu22.04 AS builder

# 安装依赖
RUN apt-get update && apt-get install -y \
    build-essential cmake git wget curl \
    libssl-dev pkg-config

# 设置 Go 环境
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOARCH=amd64
ENV GOOS=linux
ENV CC=/usr/local/occlum/bin/occlum-gcc
ENV CXX=/usr/local/occlum/bin/occlum-g++

WORKDIR /root/securemem
COPY . .

RUN occlum-go mod tidy

RUN occlum-go build -o app main.go

RUN mkdir -p occlum_instance/image/bin && \
    cp app occlum_instance/image/bin/

RUN cd occlum_instance && \
    occlum init && \
    occlum build

RUN printf '#!/bin/bash\n\
set -e\n\
echo "Starting AESM service..."\n\
export LD_LIBRARY_PATH=/opt/intel/sgx-aesm-service/aesm:/usr/lib:/opt/intel/sgxsdk/lib64:/usr/local/occlum/x86_64-linux-musl/lib:$LD_LIBRARY_PATH\n\
/opt/intel/sgx-aesm-service/aesm/aesm_service &\n\
sleep 2\n\
echo \"Running Occlum SGX App...\"\n\
cd /root/securemem/occlum_instance && exec occlum run /bin/app\n' > /start.sh && \
    chmod +x /start.sh

ENTRYPOINT ["/bin/bash", "/start.sh"]
