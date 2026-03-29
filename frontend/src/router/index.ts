import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import TabsPage from '../views/TabsPage.vue';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    component: () => import('@/views/LoginPage.vue'),
  },
  {
    path: '/tabs/',
    component: TabsPage,
    children: [
      {
        path: '',
        redirect: '/tabs/errors',
      },
      {
        path: 'errors',
        component: () => import('@/views/ErrorListPage.vue'),
      },
      {
        path: 'review',
        component: () => import('@/views/ReviewHomePage.vue'),
      },
      {
        path: 'stats',
        component: () => import('@/views/StatsPage.vue'),
      },
      {
        path: 'profile',
        component: () => import('@/views/ProfilePage.vue'),
      },
    ],
  },
  {
    path: '/records/new',
    component: () => import('@/views/ErrorEditorPage.vue'),
  },
  {
    path: '/records/:id',
    component: () => import('@/views/ErrorDetailPage.vue'),
  },
  {
    path: '/pdf/jobs/:jobId',
    component: () => import('@/views/PdfJobDetailPage.vue'),
  },
  {
    path: '/ocr/progress',
    component: () => import('@/views/OcrProgressPage.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach((to) => {
  const token = localStorage.getItem('accessToken');
  const isPublic = to.path === '/login';

  if (!token && !isPublic) {
    return '/login';
  }

  if (token && to.path === '/login') {
    return '/tabs/errors';
  }

  return true;
});

export default router;
