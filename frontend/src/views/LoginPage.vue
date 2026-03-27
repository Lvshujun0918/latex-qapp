<template>
  <section class="auth-page">
    <div class="auth-card">
      <div class="auth-head">
        <span class="auth-kicker">Latex QApp</span>
        <h1>{{ isRegisterMode ? '注册账号' : '欢迎回来' }}</h1>
        <p>{{ isRegisterMode ? '创建账号后即可同步与导出错题。' : '登录后继续你的 AI 错题本。' }}</p>
      </div>

      <div class="auth-form">
        <label class="field">
          <span>用户名</span>
          <input v-model="username" autocomplete="username" placeholder="请输入用户名" />
        </label>

        <label class="field">
          <span>密码</span>
          <input v-model="password" type="password" autocomplete="current-password" placeholder="至少 6 位密码" />
        </label>

        <label v-if="isRegisterMode" class="field">
          <span>确认密码</span>
          <input v-model="confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入密码" />
        </label>

        <p v-if="errorMessage" class="error-tip">{{ errorMessage }}</p>

        <button class="btn btn-primary" :disabled="!canSubmit || authStore.loading" @click="submit">
          {{ isRegisterMode ? '注册并登录' : '登录' }}
        </button>
        <button class="btn btn-ghost" @click="toggleMode">
          {{ isRegisterMode ? '已有账号？去登录' : '没有账号？去注册' }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

const authStore = useAuthStore();
const router = useRouter();
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const isRegisterMode = ref(false);
const errorMessage = ref('');

const canSubmit = computed(() => {
  if (username.value.trim().length <= 2 || password.value.length <= 5) {
    return false;
  }
  if (!isRegisterMode.value) {
    return true;
  }
  return confirmPassword.value === password.value;
});

async function submit() {
  errorMessage.value = '';
  const payload = { username: username.value.trim(), password: password.value };

  if (isRegisterMode.value && confirmPassword.value !== password.value) {
    errorMessage.value = '两次输入的密码不一致';
    return;
  }

  try {
    if (isRegisterMode.value) {
      await authStore.register(payload);
    }
    await authStore.login(payload);
    router.replace('/tabs/errors');
  } catch (error: any) {
    errorMessage.value = error?.response?.data?.error || error?.message || '请求失败，请重试';
  }
}

function toggleMode() {
  isRegisterMode.value = !isRegisterMode.value;
  confirmPassword.value = '';
  errorMessage.value = '';
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at 10% 10%, rgba(16, 185, 129, 0.24), transparent 38%),
    radial-gradient(circle at 88% 12%, rgba(59, 130, 246, 0.22), transparent 42%),
    radial-gradient(circle at 50% 100%, rgba(236, 72, 153, 0.12), transparent 48%),
    linear-gradient(160deg, #f8fafc 0%, #edf5ff 52%, #f0fdfa 100%);
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(148, 163, 184, 0.36);
  border-radius: 22px;
  box-shadow:
    0 24px 50px rgba(15, 23, 42, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(14px);
  padding: 24px;
  animation: auth-rise 380ms ease-out both;
}

.auth-kicker {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(16, 185, 129, 0.35);
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.auth-head h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #0f172a;
}

.auth-head p {
  margin: 10px 0 0;
  color: #475569;
  font-size: 14px;
}

.auth-form {
  margin-top: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field span {
  font-size: 13px;
  color: #334155;
  font-weight: 600;
}

.field input {
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  height: 42px;
  padding: 0 12px;
  font-size: 14px;
  background: #fff;
}

.field input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.16);
}

.btn {
  border: none;
  border-radius: 10px;
  height: 42px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(140deg, #0ea5e9 0%, #2563eb 55%, #0f766e 100%);
  color: #fff;
}

.btn-ghost {
  background: transparent;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.error-tip {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}

@keyframes auth-rise {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
