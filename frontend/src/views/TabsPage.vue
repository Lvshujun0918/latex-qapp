<template>
  <div class="tabs-layout" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <main class="tabs-content">
      <RouterView />
    </main>

    <nav class="main-tabbar" aria-label="主导航">
      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/errors') }" to="/tabs/errors">
        <BookOpen class="tab-icon" />
        <span>错题</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/review') }" to="/tabs/review">
        <GraduationCap class="tab-icon" />
        <span>复习</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/pdfs') }" to="/tabs/pdfs">
        <FileText class="tab-icon" />
        <span>题单</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/stats') }" to="/tabs/stats">
        <BarChart3 class="tab-icon" />
        <span>统计</span>
      </RouterLink>

      <RouterLink class="tab-link" :class="{ active: isActive('/tabs/profile') }" to="/tabs/profile">
        <CircleUserRound class="tab-icon" />
        <span>家长</span>
      </RouterLink>
    </nav>

  </div>
</template>

<script setup lang="ts">
import { useRoute, RouterLink, RouterView } from 'vue-router';
import { BarChart3, BookOpen, CircleUserRound, FileText, GraduationCap } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';

const route = useRoute();
const { resolvedTheme } = useTheme();

function isActive(path: string) {
  return route.path.startsWith(path);
}
</script>

<style scoped>
.tabs-layout {
  min-height: 100dvh;
  background:
    radial-gradient(circle at 16% -8%, rgba(124, 186, 255, 0.28), transparent 40%),
    radial-gradient(circle at 95% 4%, rgba(89, 219, 171, 0.2), transparent 34%),
    radial-gradient(circle at 50% 100%, rgba(255, 255, 255, 0.7), transparent 48%),
    hsl(var(--background));
}

.tabs-layout.is-dark {
  background:
    radial-gradient(circle at 16% -8%, rgba(56, 189, 248, 0.18), transparent 40%),
    radial-gradient(circle at 95% 4%, rgba(16, 185, 129, 0.16), transparent 34%),
    radial-gradient(circle at 50% 100%, rgba(15, 23, 42, 0.5), transparent 48%),
    hsl(var(--background));
}

.tabs-content {
  padding-top: env(safe-area-inset-top, 0px);
  padding-left: 16px;
  padding-right: 16px;
}

.main-tabbar {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(10px + env(safe-area-inset-bottom, 0px));
  height: 68px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  align-items: center;
  gap: 6px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.56);
  background: rgba(255, 255, 255, 0.66);
  box-shadow:
    0 -8px 30px rgba(18, 34, 56, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
  backdrop-filter: saturate(185%) blur(24px);
  padding: 0 8px;
}

.tabs-layout.is-dark .main-tabbar {
  border: 1px solid rgba(148, 163, 184, 0.26);
  background: rgba(15, 23, 42, 0.72);
  box-shadow:
    0 -8px 26px rgba(2, 6, 23, 0.44),
    inset 0 1px 0 rgba(148, 163, 184, 0.12);
}

.tab-link {
  height: 56px;
  border-radius: 14px;
  text-decoration: none;
  color: hsl(var(--muted-foreground));
  display: inline-flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  transition: transform 0.2s ease, background-color 0.2s ease, color 0.2s ease;
}

.tab-link.active {
  color: hsl(var(--primary));
  background: linear-gradient(160deg, rgba(255, 255, 255, 0.7), rgba(219, 236, 255, 0.55));
}

.tabs-layout.is-dark .tab-link {
  color: rgba(203, 213, 225, 0.9);
}

.tabs-layout.is-dark .tab-link.active {
  color: #dbeafe;
  background: linear-gradient(160deg, rgba(30, 41, 59, 0.8), rgba(37, 99, 235, 0.22));
}

.tab-icon {
  width: 18px;
  height: 18px;
}
</style>
