<template>
  <div class="profile-page">
    <div class="profile-header">
      <div class="avatar-section">
        <el-avatar :size="100" :src="profile.avatar" :key="profile.avatar" class="avatar">
          {{ (profile.username || userId || '?').charAt(0).toUpperCase() }}
        </el-avatar>
        <div class="avatar-upload">
          <el-input v-model="profile.avatar" placeholder="输入头像URL（输入后实时预览）" />
        </div>
      </div>
      <div class="user-info">
        <h2>{{ profile.username || userId }}</h2>
        <p class="user-id">ID: {{ userId }}</p>
      </div>
    </div>

    <el-card class="profile-card">
      <template #header>
        <span>基本信息</span>
      </template>
      <el-form :model="profile" label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="profile.username" placeholder="设置显示名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveProfile" :loading="saving">保存信息</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="profile-card">
      <template #header>
        <span>修改密码</span>
      </template>
      <el-form ref="pwdFormRef" :model="passwordForm" :rules="pwdRules" label-width="100px">
        <el-form-item label="当前密码" prop="old_password">
          <el-input v-model="passwordForm.old_password" type="password" show-password placeholder="输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="输入新密码" />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirm_password">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password placeholder="再次输入新密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="changePassword" :loading="changingPwd">修改密码</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="back-btn">
      <el-button @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authAPI, userAPI } from '../api'

const router = useRouter()
const saving = ref(false)
const changingPwd = ref(false)
const pwdFormRef = ref(null)

const userId = computed(() => {
  return document.cookie
    .split('; ')
    .find(row => row.startsWith('web_ssh_user='))
    ?.split('=')[1] || ''
})

const profile = reactive({
  username: '',
  avatar: ''
})

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const validateConfirm = (rule, value, callback) => {
  if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const pwdRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [{ required: true, min: 4, message: '密码至少4位', trigger: 'blur' }],
  confirm_password: [{ required: true, validator: validateConfirm, trigger: 'blur' }]
}

onMounted(async () => {
  await loadProfile()
})

const loadProfile = async () => {
  try {
    const res = await userAPI.getProfile()
    if (res.data.success) {
      profile.username = res.data.username || ''
      profile.avatar = res.data.avatar || ''
    }
  } catch (err) {
    console.error('加载用户信息失败', err)
  }
}

const saveProfile = async () => {
  saving.value = true
  try {
    const res = await userAPI.updateProfile({
      username: profile.username,
      avatar: profile.avatar
    })
    if (res.data.success) {
      ElMessage.success('信息保存成功')
    } else {
      ElMessage.error(res.data.message || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

const changePassword = async () => {
  if (!pwdFormRef.value) return

  const valid = await pwdFormRef.value.validate().catch(() => false)
  if (!valid) return

  changingPwd.value = true
  try {
    const res = await userAPI.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    if (res.data.success) {
      ElMessage.success('密码修改成功')
      passwordForm.old_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
    } else {
      ElMessage.error(res.data.message || '修改失败')
    }
  } finally {
    changingPwd.value = false
  }
}
</script>

<style scoped>
.profile-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 24px;
  height: 100%;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 24px;
  padding: 24px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.avatar {
  background: var(--accent, #3B82F6);
  font-size: 36px;
  font-weight: bold;
}

.avatar-upload {
  width: 200px;
}

.user-info h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  color: var(--text-primary);
}

.user-id {
  margin: 0;
  color: var(--text-muted);
  font-size: 14px;
}

.profile-card {
  margin-bottom: 16px;
}

.back-btn {
  text-align: center;
  margin-top: 24px;
}

@media (max-width: 640px) {
  .profile-page {
    padding: 16px;
  }

  .profile-header {
    flex-direction: column;
    text-align: center;
  }
}
</style>
