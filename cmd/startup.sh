#!/usr/bin/env bash

CMD="./go-pipeline.bin"
LOG_DIR="log"
LOG_FILE="${LOG_DIR}/logging.log"

# 备份前一份日志
mkdir -p log

# 检查Log文件
if [[ -f ${LOG_FILE} ]]; then
    BACKUP_FILE="${LOG_DIR}/"`date +"%Y%m%d%H%M%S"`".log"
    mv ${LOG_FILE} ${BACKUP_FILE}
fi

# 启动
nohup ${CMD} > ${LOG_FILE} 2>&1 &