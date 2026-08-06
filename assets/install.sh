#!/bin/bash
INSTALL_DIR="/opt/chenze-faka"
SERVICE_NAME="chenze-faka"

echo "正在安装晨泽发卡系统..."

mkdir -p $INSTALL_DIR
cp chenze_faka $INSTALL_DIR/
cp -r assets $INSTALL_DIR/
cd $INSTALL_DIR

if [ ! -f config.yaml ]; then
    cp assets/config.yaml.example config.yaml
fi

cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=Chenze FaKa System
After=network.target

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/chenze_faka
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo "晨泽发卡系统安装完成!"
echo "访问地址: http://localhost:12398"
