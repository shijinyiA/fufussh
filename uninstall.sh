#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

INSTALL_DIR="/opt/web-ssh"

clear
echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║          芙芙云 Web SSH Terminal - 卸载脚本                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

confirm_uninstall() {
    echo -e "${YELLOW}警告: 此操作将删除以下内容:${NC}"
    echo -e "  - Web SSH 服务"
    echo -e "  - 安装目录: ${INSTALL_DIR}"
    echo -e "  - 系统服务文件"
    echo ""
    echo -e "${RED}注意: 数据库数据不会被删除，如需删除请手动操作${NC}"
    echo ""

    read -p "确定要卸载吗? [y/N]: " CONFIRM
    if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}已取消卸载${NC}"
        exit 0
    fi
}

stop_services() {
    echo -e "\n${YELLOW}[1/4] 停止服务...${NC}"

    if systemctl is-active --quiet web-ssh 2>/dev/null; then
        systemctl stop web-ssh
        echo -e "${GREEN}web-ssh 服务已停止${NC}"
    else
        echo -e "${YELLOW}web-ssh 服务未运行${NC}"
    fi

    pkill -f "web-ssh" 2>/dev/null
}

disable_services() {
    echo -e "\n${YELLOW}[2/4] 禁用服务...${NC}"

    if systemctl is-enabled --quiet web-ssh 2>/dev/null; then
        systemctl disable web-ssh
        echo -e "${GREEN}web-ssh 服务已禁用${NC}"
    fi
}

remove_files() {
    echo -e "\n${YELLOW}[3/4] 删除文件...${NC}"

    if [ -f /etc/systemd/system/web-ssh.service ]; then
        rm -f /etc/systemd/system/web-ssh.service
        systemctl daemon-reload
        echo -e "${GREEN}系统服务文件已删除${NC}"
    fi

    if [ -d "$INSTALL_DIR" ]; then
        rm -rf "$INSTALL_DIR"
        echo -e "${GREEN}安装目录已删除: ${INSTALL_DIR}${NC}"
    else
        echo -e "${YELLOW}安装目录不存在${NC}"
    fi
}

show_result() {
    echo -e "\n${YELLOW}[4/4] 卸载完成${NC}"

    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                  卸载成功!                                 ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}提示:${NC}"
    echo -e "  - 数据库数据仍保留，如需删除请手动操作"
    echo -e "  - 如需重新安装，请再次运行安装脚本"
    echo ""
}

main() {
    check_root
    confirm_uninstall
    stop_services
    disable_services
    remove_files
    show_result
}

main
