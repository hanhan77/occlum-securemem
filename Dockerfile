FROM occlum/occlum:0.31.0-ubuntu20.04

RUN rm -f /etc/apt/sources.list.d/intel-sgxsdk.list
ENV GO111MODULE=on

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    g++ \
    make \
    cmake \
    git \
    wget \
    libssl-dev \
    pkg-config

# 修复 GLIBCXX_3.4.26 缺失问题
RUN cp /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /usr/local/occlum/x86_64-linux-musl/lib/

# 设置工作目录
WORKDIR /root/occlum-securemem

# 复制项目源代码（你提前准备的 go.mod/main.go/securemem/*.go）
COPY . .

# 初始化并构建 enclave 组件（模拟 seal 库）
RUN mkdir -p enclave && \
    echo "// dummy placeholder for seal" > enclave/seal.c && \
    occlum-gcc -fPIC -c enclave/seal.c -o enclave/seal.o && \
    occlum-gcc -shared -o enclave/libseal.so enclave/seal.o && \
    ar rcs enclave/libseal.a enclave/seal.o

# 初始化 Go module
RUN occlum-go mod tidy

# 编译 Go 应用（使用 occlum-go）
ENV CGO_ENABLED=1 \
    GOARCH=amd64 \
    GOOS=linux \
    CGO_CFLAGS="-I/root/occlum-securemem/enclave -Wno-error=parentheses" \
    CGO_LDFLAGS="-L/root/occlum-securemem/enclave -lseal -static-libstdc++ -static-libgcc"

RUN occlum-go build -o app main.go

# 构建 Occlum instance 镜像目录结构
RUN mkdir -p occlum_instance/image/bin && \
    mkdir -p occlum_instance/image/lib && \
    cp app occlum_instance/image/bin/ && \
    cp enclave/libseal.so occlum_instance/image/lib/ && \
    cp enclave/libseal.a occlum_instance/image/lib/ && \
    cp /usr/local/occlum/x86_64-linux-musl/lib/libstdc++.so.6 occlum_instance/image/lib/ && \
    cp /usr/local/occlum/x86_64-linux-musl/lib/libc.so occlum_instance/image/lib/

# 初始化并构建 Occlum 镜像
RUN cd occlum_instance && \
    occlum init && \
    occlum build

# 启动脚本（自动运行 AESM + occlum app）
WORKDIR /root/occlum-securemem/occlum_instance
RUN printf '#!/bin/bash\n\
set -e\n\
echo "[*] Starting AESM service..."\n\
export LD_LIBRARY_PATH=/opt/intel/sgx-aesm-service/aesm:/usr/lib:/opt/intel/sgxsdk/lib64:/usr/local/occlum/x86_64-linux-musl/lib:$LD_LIBRARY_PATH\n\
/opt/intel/sgx-aesm-service/aesm/aesm_service &\n\
sleep 2\n\
echo "[*] Running Occlum SGX App..."\n\
exec occlum run /bin/app\n' > /start.sh && chmod +x /start.sh

ENTRYPOINT ["/bin/bash", "/start.sh"]
