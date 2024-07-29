#!/usr/bin/env bash

# 一键启动该目录下所有服务

# 遍历当前目录下所有服务
for service in `ls`
do
    if [ -d "$service" ]; then
        cd "$service" || continue # 确保cd成功，否则跳过本次循环
        if [ -f "main.go" ]; then

            # 如果存在PID文件，就停止当前服务
            pid_file="$service.pid"
            if [ -f "$pid_file" ]; then
                echo "检测到服务: $service 的PID文件，因此不会启动新服务。"
                cd ..
                continue
            fi

            echo "正在后台启动服务: $service"
            nohup go run . > "$service.log" & # 后台启动服务，并将输出重定向到日志文件
            cd ..
        fi
    fi
done