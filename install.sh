#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="/opt/web-ssh"
NPM_REGISTRY="${NPM_REGISTRY:-https://registry.npmmirror.com}"

clear
echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║          芙芙云 Web SSH Terminal 一键安装脚本              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

check_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VER=$VERSION_ID
        OS_LIKE=$ID_LIKE
    elif [ -f /etc/redhat-release ]; then
        OS="centos"
    else
        OS="unknown"
    fi
    echo -e "${GREEN}检测到系统: ${OS} ${OS_VER}${NC}"
}

get_pkg_manager() {
    case $OS in
        ubuntu|debian|linuxmint|pop|elementary|kali|raspbian)
            PKG_MANAGER="apt"
            PKG_INSTALL="apt-get install -y"
            PKG_UPDATE="apt-get update -y"
            ;;
        centos|rhel|rocky|almalinux|ol|virtuozzo)
            if command -v dnf &> /dev/null; then
                PKG_MANAGER="dnf"
                PKG_INSTALL="dnf install -y"
            else
                PKG_MANAGER="yum"
                PKG_INSTALL="yum install -y"
            fi
            PKG_UPDATE="yum update -y"
            ;;
        fedora)
            PKG_MANAGER="dnf"
            PKG_INSTALL="dnf install -y"
            PKG_UPDATE="dnf update -y"
            ;;
        arch|manjaro|endeavouros|garuda|archcraft)
            PKG_MANAGER="pacman"
            PKG_INSTALL="pacman -S --noconfirm"
            PKG_UPDATE="pacman -Sy"
            ;;
        alpine)
            PKG_MANAGER="apk"
            PKG_INSTALL="apk add"
            PKG_UPDATE="apk update"
            ;;
        opensuse*|sles|sled)
            PKG_MANAGER="zypper"
            PKG_INSTALL="zypper install -y"
            PKG_UPDATE="zypper refresh"
            ;;
        void)
            PKG_MANAGER="xbps"
            PKG_INSTALL="xbps-install -y"
            PKG_UPDATE="xbps-install -Su"
            ;;
        gentoo|funtoo)
            PKG_MANAGER="emerge"
            PKG_INSTALL="emerge -v"
            PKG_UPDATE="emerge --sync"
            ;;
        slackware)
            PKG_MANAGER="slackpkg"
            PKG_INSTALL="slackpkg install"
            PKG_UPDATE="slackpkg update"
            ;;
        *)
            PKG_MANAGER="unknown"
            ;;
    esac
}

install_deps() {
    echo -e "\n${YELLOW}[1/7] 检查并安装依赖...${NC}"

    get_pkg_manager

    if [[ "$PKG_MANAGER" == "unknown" ]]; then
        echo -e "${RED}不支持的系统: $OS${NC}"
        echo -e "${YELLOW}请手动安装以下依赖: Go, Node.js, MySQL Client${NC}"
        read -p "已手动安装依赖，继续安装? [y/N]: " MANUAL_DEPS
        if [[ ! "$MANUAL_DEPS" =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        local need_go=0
        local need_node=0
        local need_mysql_client=0

        if ! command -v go &> /dev/null; then
            need_go=1
        else
            echo -e "  ${GREEN}✓${NC} Go $(go version | awk '{print $3}')"
        fi

        if ! command -v node &> /dev/null; then
            need_node=1
        else
            echo -e "  ${GREEN}✓${NC} Node.js $(node -v)"
        fi

        if ! command -v mysql &> /dev/null; then
            need_mysql_client=1
        else
            echo -e "  ${GREEN}✓${NC} MySQL Client"
        fi

        if [[ $need_go -eq 1 || $need_node -eq 1 || $need_mysql_client -eq 1 ]]; then
            echo -e "${YELLOW}正在安装缺失的依赖...${NC}"

            $PKG_UPDATE

            case $PKG_MANAGER in
                apt)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL golang-go
                    [[ $need_node -eq 1 ]] && {
                        curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
                        apt-get install -y nodejs
                    }
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql-client
                    ;;
                dnf|yum)
                    [[ "$PKG_MANAGER" == "yum" ]] && yum install -y epel-release
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL golang
                    [[ $need_node -eq 1 ]] && {
                        curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
                        $PKG_INSTALL nodejs
                    }
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql
                    ;;
                pacman)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL nodejs npm
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql-clients
                    ;;
                apk)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL nodejs npm
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql-client
                    ;;
                zypper)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL nodejs
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql-client
                    ;;
                xbps)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL nodejs
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql-client
                    ;;
                emerge)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL dev-lang/go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL net-libs/nodejs
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL dev-db/mysql
                    ;;
                slackpkg)
                    [[ $need_go -eq 1 ]] && $PKG_INSTALL go
                    [[ $need_node -eq 1 ]] && $PKG_INSTALL node
                    [[ $need_mysql_client -eq 1 ]] && $PKG_INSTALL mysql
                    ;;
            esac
        fi
    fi

    echo -e "${GREEN}依赖检查完成${NC}"
}

get_user_input() {
    echo -e "\n${YELLOW}[2/7] 配置信息输入${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    echo -e "\n${BLUE}【MySQL 数据库配置】${NC}"
    echo -e "${YELLOW}提示: 请先在宝塔面板中创建数据库${NC}"
    echo -e "  - 数据库名称: ${GREEN}ssh-web${NC} (必须是这个名称)"
    echo -e "  - 用户名: 建议与数据库名相同"
    echo -e "  - 访问权限: 选择"所有人"或指定服务器IP"
    echo ""

    read -p "请输入数据库服务器 IP: " MYSQL_HOST
    while [[ -z "$MYSQL_HOST" ]]; do
        echo -e "${RED}数据库 IP 不能为空${NC}"
        read -p "请输入数据库服务器 IP: " MYSQL_HOST
    done

    read -p "请输入数据库端口 [默认 3306]: " MYSQL_PORT
    MYSQL_PORT=${MYSQL_PORT:-3306}

    read -p "请输入数据库用户名: " MYSQL_USER
    while [[ -z "$MYSQL_USER" ]]; do
        echo -e "${RED}数据库用户名不能为空${NC}"
        read -p "请输入数据库用户名: " MYSQL_USER
    done

    read -s -p "请输入数据库密码: " MYSQL_PASS
    echo
    while [[ -z "$MYSQL_PASS" ]]; do
        echo -e "${RED}数据库密码不能为空${NC}"
        read -s -p "请输入数据库密码: " MYSQL_PASS
        echo
    done

    MYSQL_DB="ssh-web"

    echo -e "\n${BLUE}【网站外观配置】${NC}"
    read -p "请输入网站名称 [默认 芙芙云]: " SITE_NAME
    SITE_NAME=${SITE_NAME:-芙芙云}

    read -p "请输入登录页背景图片 URL [默认 https://www.loliapi.com/acg/]: " BG_IMAGE
    BG_IMAGE=${BG_IMAGE:-https://www.loliapi.com/acg/}

    read -p "请输入网站 Logo 图片 URL: " LOGO_URL
    while [[ -z "$LOGO_URL" ]]; do
        echo -e "${RED}Logo URL 不能为空${NC}"
        read -p "请输入网站 Logo 图片 URL: " LOGO_URL
    done

    echo -e "\n${BLUE}【服务端口配置】${NC}"
    read -p "请输入服务端口 [默认 8080]: " SERVER_PORT
    SERVER_PORT=${SERVER_PORT:-8080}

    echo -e "\n${BLUE}【系统服务配置】${NC}"
    read -p "是否创建系统服务并设置开机自启? [Y/n]: " CREATE_SERVICE
    CREATE_SERVICE=${CREATE_SERVICE:-Y}

    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}配置信息确认:${NC}"
    echo -e "  MySQL: ${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB} (用户: ${MYSQL_USER})"
    echo -e "  网站名称: ${SITE_NAME}"
    echo -e "  背景图片: ${BG_IMAGE}"
    echo -e "  Logo: ${LOGO_URL}"
    echo -e "  服务端口: ${SERVER_PORT}"
    echo -e "  管理员后台: http://你的IP:${SERVER_PORT}/admin"
    echo -e "  系统服务: $([[ "$CREATE_SERVICE" =~ ^[Yy]$ ]] && echo "${GREEN}是${NC}" || echo "${YELLOW}否${NC}")"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    read -p "确认以上信息正确? [Y/n]: " CONFIRM
    CONFIRM=${CONFIRM:-Y}
    if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
        echo -e "${RED}用户取消安装${NC}"
        exit 0
    fi
}

test_mysql() {
    echo -e "\n${YELLOW}[3/7] 测试数据库连接...${NC}"

    local MYSQL_CLIENT="mysql"
    local mysql_output
    mysql_output=$(mysql -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" -p"$MYSQL_PASS" -e "USE \`$MYSQL_DB\`;" 2>&1)
    local ret=$?

    if [[ $ret -eq 0 ]]; then
        echo -e "${GREEN}数据库连接成功${NC}"
        return
    fi

    if echo "$mysql_output" | grep -qi "caching_sha2_password"; then
        echo -e "${YELLOW}检测到 MySQL 认证插件问题，尝试使用 Docker 容器内的 MySQL 客户端修复...${NC}"

        local DOCKER_MYSQL_CONTAINER
        DOCKER_MYSQL_CONTAINER=$(docker ps --format '{{.Names}}' 2>/dev/null | grep -i mysql | head -1)

        if [[ -z "$DOCKER_MYSQL_CONTAINER" ]]; then
            DOCKER_MYSQL_CONTAINER=$(docker ps --format '{{.Image}} {{.Names}}' 2>/dev/null | grep -i mysql | head -1 | awk '{print $2}')
        fi

        if [[ -n "$DOCKER_MYSQL_CONTAINER" ]]; then
            echo -e "${GREEN}发现 MySQL Docker 容器: ${DOCKER_MYSQL_CONTAINER}${NC}"

            if docker exec "$DOCKER_MYSQL_CONTAINER" mysql -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" -p"$MYSQL_PASS" \
                -e "ALTER USER '${MYSQL_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}'; FLUSH PRIVILEGES;" 2>/dev/null; then
                echo -e "${GREEN}用户认证方式已切换为 mysql_native_password${NC}"
            elif docker exec "$DOCKER_MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" \
                -e "ALTER USER '${MYSQL_USER}'@'%' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}'; FLUSH PRIVILEGES;" 2>/dev/null; then
                echo -e "${GREEN}用户认证方式已切换为 mysql_native_password${NC}"
            else
                echo -e "${YELLOW}自动修复失败，请手动执行以下命令（在宿主机或 Docker 容器内）：${NC}"
                echo -e "  ${CYAN}ALTER USER '${MYSQL_USER}'@'%' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}';${NC}"
                echo -e "  ${CYAN}ALTER USER '${MYSQL_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}';${NC}"
                echo -e "  ${CYAN}FLUSH PRIVILEGES;${NC}"
                echo ""
                echo -e "快速修复命令："
                echo -e "  ${CYAN}docker exec ${DOCKER_MYSQL_CONTAINER} mysql -uroot -p${NC}"
                echo -e "  然后在 MySQL 提示符中执行上面的 ALTER USER 语句"
                exit 1
            fi
        else
            echo -e "${YELLOW}未找到 MySQL Docker 容器，请手动修复：${NC}"
            echo -e "进入 MySQL Docker 容器执行以下 SQL："
            echo -e "  ${CYAN}docker exec -it <容器名> mysql -uroot -p${NC}"
            echo -e "  然后执行："
            echo -e "  ${CYAN}ALTER USER '${MYSQL_USER}'@'%' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}';${NC}"
            echo -e "  ${CYAN}ALTER USER '${MYSQL_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '${MYSQL_PASS}';${NC}"
            echo -e "  ${CYAN}FLUSH PRIVILEGES;${NC}"
            exit 1
        fi

        if mysql -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" -p"$MYSQL_PASS" -e "USE \`$MYSQL_DB\`;" 2>&1; then
            echo -e "${GREEN}数据库连接成功（认证方式已切换为 mysql_native_password）${NC}"
        else
            echo -e "${RED}数据库连接失败，请检查:${NC}"
            echo -e "  1. 是否已在 Docker MySQL 中创建数据库 ${GREEN}ssh-web${NC}"
            echo -e "  2. 数据库用户名和密码是否正确"
            echo -e "  3. 数据库访问权限是否允许当前服务器IP"
            exit 1
        fi
    else
        echo -e "${RED}数据库连接失败，原因：${NC}"
        echo "$mysql_output"
        echo -e "请检查:"
        echo -e "  1. 是否已在 Docker MySQL 中创建数据库 ${GREEN}ssh-web${NC}"
        echo -e "  2. 数据库用户名和密码是否正确"
        echo -e "  3. 数据库访问权限是否允许当前服务器IP"
        exit 1
    fi
}

copy_files() {
    echo -e "\n${YELLOW}[4/7] 复制文件到安装目录...${NC}"

    mkdir -p "$INSTALL_DIR"

    rm -rf "$INSTALL_DIR/frontend" "$INSTALL_DIR/backend"

    cp -r "$SCRIPT_DIR/frontend" "$INSTALL_DIR/"
    cp -r "$SCRIPT_DIR/backend" "$INSTALL_DIR/"
    rm -f "$INSTALL_DIR/backend/web-ssh" "$INSTALL_DIR/backend/admin"

    if [[ -f "$INSTALL_DIR/backend/admin_dist/index.html" ]]; then
        echo -e "${GREEN}✓${NC} 管理员后台预编译产物已就绪"
    fi

    echo -e "${GREEN}文件复制完成${NC}"
}

build_ssh_web() {
    echo -e "\n${YELLOW}[5/7] 编译 SSH Web 前端和后端...${NC}"

    MYSQL_DSN="${MYSQL_USER}:${MYSQL_PASS}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DB}?charset=utf8mb4&parseTime=True&loc=Local"

    cat > "$INSTALL_DIR/backend/.env" << EOF
PORT=${SERVER_PORT}
MYSQL_DSN=${MYSQL_DSN}
EOF

    cat > "$INSTALL_DIR/backend/config.json" << EOF
{
    "mysql_host": "${MYSQL_HOST}",
    "mysql_port": ${MYSQL_PORT},
    "mysql_user": "${MYSQL_USER}",
    "mysql_pass": "${MYSQL_PASS}",
    "mysql_db": "${MYSQL_DB}",
    "admin_port": "${SERVER_PORT}"
}
EOF

    sed -i "s|<title>.*</title>|<title>${SITE_NAME} SSH Terminal</title>|g" "$INSTALL_DIR/frontend/index.html"

    sed -i "s|src=\"https://www.loliapi.com/acg/\"|src=\"${BG_IMAGE}\"|g" "$INSTALL_DIR/frontend/src/views/Login.vue"
    sed -i "s|src=\"https://logo.fufuidc.com/logo3.png\"|src=\"${LOGO_URL}\"|g" "$INSTALL_DIR/frontend/src/views/Login.vue"
    sed -i "s|alt=\"芙芙云\"|alt=\"${SITE_NAME}\"|g" "$INSTALL_DIR/frontend/src/views/Login.vue"
    sed -i "s|class=\"logo-text\">芙芙云|class=\"logo-text\">${SITE_NAME}|g" "$INSTALL_DIR/frontend/src/views/Login.vue"

    echo -e "${CYAN}正在编译前端...${NC}"
    cd "$INSTALL_DIR/frontend"

    if [ -f package-lock.json ]; then
        npm ci --registry="${NPM_REGISTRY}" --no-audit --prefer-offline
    else
        npm install --registry="${NPM_REGISTRY}" --no-audit --prefer-offline
    fi

    npm run build

    if [ $? -ne 0 ]; then
        echo -e "${RED}前端编译失败${NC}"
        exit 1
    fi

    echo -e "${GREEN}前端编译完成${NC}"

    echo -e "${CYAN}正在编译后端...${NC}"
    cd "$INSTALL_DIR/backend"

    export GOPROXY=https://goproxy.cn,direct
    go mod download
    go build -o web-ssh

    if [ $? -ne 0 ]; then
        echo -e "${RED}后端编译失败${NC}"
        exit 1
    fi

    echo -e "${GREEN}后端编译完成${NC}"
}

setup_admin() {
    echo -e "\n${YELLOW}[6/7] 配置管理员后台...${NC}"

    if [[ -f "$INSTALL_DIR/backend/admin_dist/index.html" ]]; then
        echo -e "${GREEN}管理员后台配置完成: 与 Web SSH 共用 ${SERVER_PORT} 端口，通过 /admin 访问${NC}"
        echo -e "${YELLOW}首次访问请打开 http://你的服务器IP:${SERVER_PORT}/admin 初始化管理员账户${NC}"
    else
        echo -e "${RED}未找到管理员后台预编译产物 backend/admin_dist，请使用完整发布包安装${NC}"
        exit 1
    fi
}

setup_service() {
    if [[ "$CREATE_SERVICE" =~ ^[Yy]$ ]]; then
        echo -e "\n${YELLOW}[7/7] 创建系统服务...${NC}"

        cat > /etc/systemd/system/web-ssh.service << EOF
[Unit]
Description=Web SSH Terminal Service
After=network.target mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}/backend
EnvironmentFile=${INSTALL_DIR}/backend/.env
ExecStart=${INSTALL_DIR}/backend/web-ssh
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

        systemctl daemon-reload
        systemctl enable web-ssh
        systemctl start web-ssh

        echo -e "${GREEN}系统服务创建完成${NC}"
    else
        echo -e "\n${YELLOW}[7/7] 跳过创建系统服务...${NC}"

        cat > "$INSTALL_DIR/start.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")/backend"
if [ -f .env ]; then
    while IFS='=' read -r key value; do
        [[ "$key" =~ ^#.*$ || -z "$key" ]] && continue
        export "$key=$value"
    done < .env
fi
./web-ssh
EOF
        chmod +x "$INSTALL_DIR/start.sh"

        cat > "$INSTALL_DIR/stop.sh" << 'EOF'
#!/bin/bash
pkill -f "web-ssh"
echo "服务已停止"
EOF
        chmod +x "$INSTALL_DIR/stop.sh"

        echo -e "${GREEN}已创建启动/停止脚本${NC}"
    fi
}

show_result() {
    clear
    echo -e "${GREEN}"
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                  安装完成!                                 ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    if [[ "$CREATE_SERVICE" =~ ^[Yy]$ ]]; then
        echo -e "${CYAN}Web SSH 服务:${NC}"
        echo -e "  启动: ${GREEN}systemctl start web-ssh${NC}"
        echo -e "  停止: ${GREEN}systemctl stop web-ssh${NC}"
        echo -e "  重启: ${GREEN}systemctl restart web-ssh${NC}"
        echo -e "  状态: ${GREEN}systemctl status web-ssh${NC}"
        echo -e "  日志: ${GREEN}journalctl -u web-ssh -f${NC}"
    else
        echo -e "${CYAN}Web SSH 服务:${NC}"
        echo -e "  启动: ${GREEN}${INSTALL_DIR}/start.sh${NC}"
        echo -e "  停止: ${GREEN}${INSTALL_DIR}/stop.sh${NC}"
        echo -e "  ${YELLOW}提示: 未创建系统服务，需手动启动${NC}"
    fi

    echo ""
    echo -e "${CYAN}访问地址:${NC}"
    echo -e "  用户登录: ${GREEN}http://你的服务器IP:${SERVER_PORT}${NC}"
    echo -e "  管理员后台: ${GREEN}http://你的服务器IP:${SERVER_PORT}/admin${NC}"
    echo -e "  ${YELLOW}首次访问管理员后台请先初始化: http://你的服务器IP:${SERVER_PORT}/admin/init${NC}"
    echo ""
    echo -e "${CYAN}安装目录:${NC}"
    echo -e "  ${GREEN}${INSTALL_DIR}${NC}"
    echo ""
    echo -e "${CYAN}卸载:${NC}"
    echo -e "  运行: ${GREEN}${SCRIPT_DIR}/uninstall.sh${NC}"
    echo ""

    if [[ "$CREATE_SERVICE" =~ ^[Yy]$ ]]; then
        systemctl status web-ssh --no-pager -l
    fi
}

main() {
    check_root
    check_os
    install_deps
    get_user_input
    test_mysql
    copy_files
    build_ssh_web
    setup_admin
    setup_service
    show_result
}

main
