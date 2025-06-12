#!/bin/bash

set -euo pipefail

# 配置变量
readonly BACKUP_DIR="${BACKUP_DIR:-/backups/backup}"
readonly DATA_DIR="${DATA_DIR:-/data}"
readonly NOW="$(date +%Y%m%d_%H%M%S)"
readonly RETENTION_DAYS="${RETENTION_DAYS:-30}"

# 启用 nullglob 防止通配符展开问题
shopt -s nullglob

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

# 清理函数
cleanup() {
    local exit_code=$?
    if [[ -d "$BACKUP_DIR" ]]; then
        color yellow "清理临时目录: $BACKUP_DIR"
        rm -rf "$BACKUP_DIR"
    fi
    if [[ $exit_code -ne 0 ]]; then
        color red "备份过程中发生错误，已清理临时文件"
    fi
    exit $exit_code
}

# 注册清理函数
trap cleanup EXIT

# 验证密码环境变量
validate_password() {
    if [[ -z "${ZIP_PASSWORD:-}" ]]; then
        color red "错误：未设置 ZIP_PASSWORD 环境变量"
        color red "请设置备份密码：export ZIP_PASSWORD='your_password'"
        return 1
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
        return 1
    fi
    
    # 验证密码环境变量
    validate_password
}

# 验证目录安全性
validate_directories() {
    # 验证备份目录路径安全性
    if [[ -z "$BACKUP_DIR" || "$BACKUP_DIR" == "/" || "$BACKUP_DIR" == "/tmp" ]]; then
        color red "备份目录路径不安全: $BACKUP_DIR"
        return 1
    fi

    # 验证数据目录存在性
    if [[ ! -d "$DATA_DIR" ]]; then
        color red "数据目录不存在: $DATA_DIR"
        return 1
    fi
}

# 初始化备份环境
init_backup() {
    validate_directories

    # 安全地清理并创建备份目录
    if [[ -d "$BACKUP_DIR" ]]; then
        color yellow "清理现有备份目录: $BACKUP_DIR"
        rm -rf "$BACKUP_DIR"
    fi
    
    mkdir -p "$BACKUP_DIR"
    color blue "备份初始化完成，时间戳: $NOW"
}

# 备份数据库
backup_database() {
    local db_file="${DATA_DIR}/db.sqlite3"
    local backup_db="${BACKUP_DIR}/db.sqlite3"

    if [[ -f "$db_file" ]]; then
        color blue "备份数据库..."
        
        # 检查数据库文件是否可读
        if [[ ! -r "$db_file" ]]; then
            color red "数据库文件无法读取: $db_file"
            return 1
        fi

        # 执行数据库备份
        if sqlite3 "$db_file" ".backup '$backup_db'"; then
            color green "数据库备份完成"
            return 0
        else
            color red "数据库备份失败"
            return 1
        fi
    else
        color yellow "数据库文件不存在，跳过"
        return 0
    fi
}

# 安全复制文件
safe_copy() {
    local src="$1"
    local dest="$2"
    local desc="${3:-文件}"

    if [[ -f "$src" ]]; then
        color blue "备份${desc}..."
        if cp "$src" "$dest"; then
            color green "${desc}备份完成"
            return 0
        else
            color red "${desc}备份失败"
            return 1
        fi
    else
        color yellow "${desc}不存在，跳过"
        return 0
    fi
}

# 备份配置文件
backup_config() {
    safe_copy "${DATA_DIR}/config.json" "${BACKUP_DIR}/config.json" "配置文件（config.json）"
}

backup_env() {
    safe_copy "${DATA_DIR}/.env" "${BACKUP_DIR}/.env" "配置文件（.env）"
}

# 备份RSA密钥
backup_rsa_keys() {
    local rsa_files=("${DATA_DIR}"/rsa_key*)
    
    if [[ ${#rsa_files[@]} -gt 0 && -f "${rsa_files[0]}" ]]; then
        color blue "备份RSA密钥..."
        local success=true
        
        for file in "${rsa_files[@]}"; do
            if [[ -f "$file" ]]; then
                if ! cp "$file" "$BACKUP_DIR/"; then
                    color red "RSA密钥备份失败: $(basename "$file")"
                    success=false
                fi
            fi
        done
        
        if $success; then
            color green "RSA密钥备份完成 (共 ${#rsa_files[@]} 个文件)"
        else
            return 1
        fi
    else
        color yellow "RSA密钥不存在，跳过"
    fi
}

# 安全复制目录
safe_copy_dir() {
    local src_dir="$1"
    local backup_dir="$2"
    local desc="$3"

    if [[ -d "$src_dir" && -n "$(ls -A "$src_dir" 2>/dev/null)" ]]; then
        color blue "备份${desc}..."
        
        if cp -rp "$src_dir" "$backup_dir/"; then
            local count=$(find "$backup_dir/$(basename "$src_dir")" -type f 2>/dev/null | wc -l)
            color green "${desc}备份完成 (共 $count 个文件)"
            return 0
        else
            color red "${desc}备份失败"
            return 1
        fi
    else
        color yellow "${desc}目录不存在或为空，跳过"
        return 0
    fi
}

# 备份附件
backup_attachments() {
    safe_copy_dir "${DATA_DIR}/attachments" "$BACKUP_DIR" "附件（attachments）"
}

# 备份发送文件
backup_sends() {
    safe_copy_dir "${DATA_DIR}/sends" "$BACKUP_DIR" "发送文件（sends）"
}

# 执行所有备份任务
run_backup() {
    color blue "开始备份任务"
    local backup_success=true

    # 执行各项备份任务，记录失败状态
    backup_database || backup_success=false
    backup_config || backup_success=false
    backup_env || backup_success=false
    backup_rsa_keys || backup_success=false
    backup_attachments || backup_success=false
    backup_sends || backup_success=false

    if $backup_success; then
        color green "所有备份任务完成"
        return 0
    else
        color yellow "备份任务完成，但部分项目失败"
        return 1
    fi
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
    local password="${ZIP_PASSWORD}"

    # 创建加密压缩包（直接使用环境变量密码）
    if tar -czf - -C "$BACKUP_DIR" . | openssl enc -aes-256-cbc -salt -pbkdf2 -pass pass:"$password" -out "$archive_file"; then
        color blue "备份压缩包: $archive_file"
        
        # 显示文件大小
        if command -v du >/dev/null 2>&1; then
            local size=$(du -h "$archive_file" | cut -f1)
            color blue "压缩包大小: $size"
        fi

        color green "压缩包创建完成"
        return 0
    else
        color red "压缩包创建失败"
        return 1
    fi
}

# 清理旧备份文件
cleanup_old_backups() {
    local backup_parent_dir
    backup_parent_dir=$(dirname "$BACKUP_DIR")
    
    if [[ -d "$backup_parent_dir" && $RETENTION_DAYS -gt 0 ]]; then
        color blue "清理 $RETENTION_DAYS 天前的备份文件..."
        
        # 查找并删除旧的备份文件
        local old_files
        old_files=$(find "$backup_parent_dir" -name "$(basename "$BACKUP_DIR")_*.tar.gz" -type f -mtime +$RETENTION_DAYS 2>/dev/null || true)
        
        if [[ -n "$old_files" ]]; then
            local count=0
            while IFS= read -r file; do
                if [[ -f "$file" ]]; then
                    rm "$file" && ((count++))
                    color yellow "已删除旧备份: $(basename "$file")"
                fi
            done <<< "$old_files"
            
            if [[ $count -gt 0 ]]; then
                color green "已清理 $count 个旧备份文件"
            fi
        else
            color blue "没有找到需要清理的旧备份文件"
        fi
    fi
}

# 主函数
main() {
    local start_time=$(date +%s)
    local overall_success=true

    color green "==================== 备份开始 ===================="

    # 执行备份流程
    check_dependencies || overall_success=false
    init_backup || overall_success=false
    run_backup || overall_success=false
    create_archive || overall_success=false
    cleanup_old_backups

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    color blue "用时: ${duration} 秒"
    
    if $overall_success; then
        color green "==================== 备份完成 ===================="
        return 0
    else
        color red "==================== 备份完成（有错误） ===================="
        return 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi