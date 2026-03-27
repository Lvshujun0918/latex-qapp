import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { apiClient } from '@/services/api';

interface LoginPayload {
  username: string;
  password: string;
}

interface RegisterPayload {
  username: string;
  password: string;
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('accessToken') || '');
  const username = ref(localStorage.getItem('username') || '');
  const displayName = ref(localStorage.getItem('displayName') || '');
  const userId = ref<number | null>(null);
  const loading = ref(false);

  const isAuthed = computed(() => !!accessToken.value);

  async function login(payload: LoginPayload) {
  loading.value = true;
  try {
    const { data } = await apiClient.post('/api/auth/login', payload);
    const authData = data?.data ?? {};
    accessToken.value = authData.accessToken ?? '';
    username.value = authData.user?.username ?? payload.username;
    displayName.value = authData.user?.display_name ?? username.value;
    userId.value = authData.user?.id ?? null;

    if (!accessToken.value) {
      throw new Error('登录失败：未返回 accessToken');
    }

    localStorage.setItem('accessToken', accessToken.value);
    localStorage.setItem('username', username.value);
    localStorage.setItem('displayName', displayName.value);
  } finally {
    loading.value = false;
  }
  }

  async function register(payload: RegisterPayload) {
  loading.value = true;
  try {
    await apiClient.post('/api/auth/register', payload);
  } finally {
    loading.value = false;
  }
  }

  async function fetchMe() {
  if (!accessToken.value) {
    return;
  }
  try {
    const { data } = await apiClient.get('/api/auth/me');
    const me = data?.data;
    if (!me) {
      return;
    }
    username.value = me.username ?? username.value;
    displayName.value = me.display_name ?? displayName.value;
    userId.value = me.id ?? userId.value;
    localStorage.setItem('username', username.value);
    localStorage.setItem('displayName', displayName.value);
  } catch {
    logout();
  }
  }

  function logout() {
    accessToken.value = '';
    username.value = '';
  displayName.value = '';
  userId.value = null;
    localStorage.removeItem('accessToken');
    localStorage.removeItem('username');
  localStorage.removeItem('displayName');
  }

  return {
    accessToken,
    username,
  displayName,
  userId,
  loading,
    isAuthed,
    login,
  register,
  fetchMe,
    logout,
  };
});
