#!/bin/bash

set -euo pipefail

# 配置变量
readonly BACKUP_DIR="${BACKUP_DIR:-/backups/backup}"
readonly DATA_DIR="${DATA_DIR:-/data}"
readonly NOW="$(date +%Y%m%d_%H%M%S)"

# 颜色输出
color() {
    local color_code=""
    case $1 in
        red) color_code="\033[31m" ;;
        green) color_code="\033[32m" ;;
        blue) color_code="\033[34m" ;;
        yellow) color_code="\033[33m" ;;
        *) color_code="" ;;
    esac

    if [[ -n "$color_code" ]]; then
        echo -e "${color_code}$2\033[0m"
    else
        echo "$2"
    fi
}

# 生成随机密码
generate_password() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 16
    else
        color red "错误：未找到 openssl 。无法生成安全密码。"
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    local deps=("sqlite3" "tar" "openssl")
    local missing=()

    for dep in "${deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing+=("$dep")
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        color red "缺少依赖: ${missing[*]}"
        exit 1
    fi
}

# 初始化备份环境
init_backup() {
    # 清理并创建备份目录
    [[ -d "$BACKUP_DIR" ]] && rm -rf "$BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"

    # 检查数据目录
    if [[ ! -d "$DATA_DIR" ]]; then
        color red "数据目录不存在: $DATA_DIR"
        exit 1
    fi

    color blue "备份初始化完成，时间戳: $NOW"
}

# 备份数据库
backup_database() {
    local db_file="${DATA_DIR}/db.sqlite3"
    local backup_db="${BACKUP_DIR}/db.sqlite3"

    if [[ -f "$db_file" ]]; then
        color blue "备份数据库..."
        sqlite3 "$db_file" ".backup '$backup_db'" || {
            color red "数据库备份失败"
            return 1
        }
        color green "数据库备份完成"
    else
        color yellow "数据库文件不存在，跳过"
    fi
}

# 备份配置文件
backup_config() {
    local config_file="${DATA_DIR}/config.json"
    local backup_config="${BACKUP_DIR}/config.json"

    if [[ -f "$config_file" ]]; then
        color blue "备份配置文件（config.json）..."
        cp "$config_file" "$backup_config"
        color green "配置文件（config.json）备份完成"
    else
        color yellow "配置文件（config.json）不存在，跳过"
    fi
}

backup_env() {
    local config_file="${DATA_DIR}/.env"
    local backup_config="${BACKUP_DIR}/.env"

    if [[ -f "$config_file" ]]; then
        color blue "备份配置文件（.env）..."
        cp "$config_file" "$backup_config"
        color green "配置文件（.env）备份完成"
    else
        color yellow "配置文件（.env）不存在，跳过"
    fi
}

# 备份RSA密钥
backup_rsa_keys() {
    local rsa_pattern="${DATA_DIR}/rsa_key*"

    if ls $rsa_pattern 1> /dev/null 2>&1; then
        color blue "备份RSA密钥..."
        cp ${rsa_pattern} "$BACKUP_DIR/" 2>/dev/null || true
        color green "RSA密钥备份完成"
    else
        color yellow "RSA密钥不存在，跳过"
    fi
}

# 备份附件
backup_attachments() {
    local attachments_dir="${DATA_DIR}/attachments"

    if [[ -d "$attachments_dir" && -n "$(ls -A "$attachments_dir" 2>/dev/null)" ]]; then
        color blue "备份附件（attachments）..."
        cp -rp "$attachments_dir" "$BACKUP_DIR/"
        local count=$(find "$BACKUP_DIR/attachments" -type f | wc -l)
        color green "附件（attachments）备份完成 (共 $count 个文件)"
    else
        color yellow "附件（attachments）目录不存在或为空，跳过"
    fi
}

# 备份发送文件
backup_sends() {
    local sends_dir="${DATA_DIR}/sends"

    if [[ -d "$sends_dir" && -n "$(ls -A "$sends_dir" 2>/dev/null)" ]]; then
        color blue "备份发送文件（sends）..."
        cp -rp "$sends_dir" "$BACKUP_DIR/"
        local count=$(find "$BACKUP_DIR/sends" -type f | wc -l)
        color green "发送文件（sends）备份完成 (共 $count 个文件)"
    else
        color yellow "发送文件（sends）目录不存在或为空，跳过"
    fi
}

# 执行所有备份任务
run_backup() {
    color blue "开始备份任务"

    backup_database
    backup_config
    backup_env
    backup_rsa_keys
    backup_attachments
    backup_sends

    color blue "备份任务完成"
}

# 打包压缩
create_archive() {
    # 检查备份目录是否有内容
    if [[ -z "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]]; then
        color yellow "备份目录为空，跳过打包"
        return 0
    fi

    color blue "创建加密压缩包..."

    local archive_file="${BACKUP_DIR}_${NOW}.tar.gz"
    local password="${ZIP_PASSWORD:-$(generate_password)}"

    # 创建加密压缩包
    if tar -czf - -C "$BACKUP_DIR" . | openssl enc -aes-256-cbc -salt -pbkdf2 -pass pass:"$password" -out "$archive_file"; then
        # 保存密码提示

        color yellow "密码: $password"
        color blue "备份压缩包: $archive_file"

        # 删除备份目录
        rm -rf "$BACKUP_DIR"
        color blue "已删除临时备份目录 $BACKUP_DIR"
    else
        color red "压缩包创建失败"
        return 1
    fi
}

# 主函数
main() {
    local start_time=$(date +%s)

    color green "==================== 备份开始 ===================="

    check_dependencies
    init_backup
    run_backup
    create_archive

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    color green "==================== 备份完成 ===================="
    color green "用时: ${duration} 秒"
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
