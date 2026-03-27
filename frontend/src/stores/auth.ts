import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

interface LoginPayload {
  username: string;
  password: string;
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(localStorage.getItem('accessToken') || '');
  const refreshToken = ref(localStorage.getItem('refreshToken') || '');
  const username = ref(localStorage.getItem('username') || '');

  const isAuthed = computed(() => !!accessToken.value);

  async function login(payload: LoginPayload) {
    accessToken.value = `mock-access-${Date.now()}`;
    refreshToken.value = `mock-refresh-${Date.now()}`;
    username.value = payload.username;
    localStorage.setItem('accessToken', accessToken.value);
    localStorage.setItem('refreshToken', refreshToken.value);
    localStorage.setItem('username', username.value);
  }

  function logout() {
    accessToken.value = '';
    refreshToken.value = '';
    username.value = '';
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('username');
  }

  return {
    accessToken,
    refreshToken,
    username,
    isAuthed,
    login,
    logout,
  };
});
