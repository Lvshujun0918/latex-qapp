<template>
    <section class="app-page app-inner-page pt-8 about-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
        <header class="app-page-header page-header">
            <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级">< </Button>
            <h1>关于应用</h1>
            <p>了解产品定位、版本信息和数据处理方式。</p>
        </header>

        <Card class="app-page-shell about-hero">
            <CardContent class="hero-content">
                <div class="hero-mark">LQ</div>
                <div class="hero-text">
                    <h2>{{ appName }}</h2>
                    <p>面向家长的孩子错题整理、回顾与强化练习工具。</p>
                </div>
            </CardContent>
        </Card>

        <div class="info-row">
            <span>版本号</span>
            <strong>{{ appVersion }}</strong>
        </div>
        <div class="info-row">
            <span>构建号</span>
            <strong>{{ appBuild }}</strong>
        </div>
        <div class="info-row">
            <span>平台</span>
            <strong>{{ runtimePlatform }}</strong>
        </div>

        <Card class="app-soft-card info-card">
            <CardHeader>
                <CardTitle>关于本应用</CardTitle>
            </CardHeader>
            <CardContent>
                面向家长的智能伴学错题本。支持拍照识别题目、LaTeX 草稿生成、AI 解题解析、错题管理与统计、按题目导出 PDF。由LaTeX与AI视觉强力驱动，0题库却能较好完成错题整理与回顾，帮助孩子高效复习、查漏补缺。我们专注于错题处理流程的智能化和易用性，致力于成为家长的得力工具，陪伴孩子的学习成长。
            </CardContent>
        </Card>
        
        <p class="copyright">Make with ❤️ by Lsj</p>

    </section>
</template>

<script setup lang="ts">
import { App } from '@capacitor/app';
import { Capacitor } from '@capacitor/core';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from '@/composables/useTheme';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

const router = useRouter();
const { resolvedTheme } = useTheme();

const appName = ref('LaTeX 错题本');
const appVersion = ref('0.0.1');
const appBuild = ref('--');
const runtimePlatform = ref('Web');

onMounted(async () => {
    runtimePlatform.value = Capacitor.getPlatform();

    try {
        const info = await App.getInfo();
        appName.value = info.name || appName.value;
        appVersion.value = info.version || appVersion.value;
        appBuild.value = info.build || appBuild.value;
    } catch {
        // Keep fallback values when platform plugin is unavailable.
    }
});

function goBack() {
    if (window.history.length > 1) {
        router.back();
        return;
    }

    router.replace('/tabs/profile');
}
</script>

<style scoped>
.about-page {
    gap: 12px;
}

.about-hero {
    overflow: hidden;
}

.hero-content {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 0;
}

.hero-mark {
    width: 54px;
    height: 54px;
    border-radius: 16px;
    display: grid;
    place-items: center;
    font-size: 20px;
    font-weight: 800;
    color: #ffffff;
    background: #2563eb;
    box-shadow: 0 10px 24px rgba(37, 99, 235, 0.24);
}

.hero-text h2 {
    margin: 0;
    font-size: 20px;
    line-height: 1.1;
    color: #0f172a;
}

.hero-text p {
    margin: 6px 0 0;
    font-size: 13px;
    color: #64748b;
}

.info-card {
    border-radius: 16px;
}

.info-list {
    display: grid;
    gap: 8px;
}

.info-row {
    border: 1px solid rgba(148, 163, 184, 0.22);
    border-radius: 12px;
    background: rgba(248, 250, 252, 0.74);
    padding: 10px 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
}

.info-row span {
    font-size: 12px;
    color: #64748b;
}

.info-row strong {
    font-size: 14px;
    color: #0f172a;
}

.about-list {
    margin: 0;
    padding-left: 18px;
    display: grid;
    gap: 8px;
    color: #334155;
    font-size: 14px;
}

.policy-text {
    display: grid;
    gap: 6px;
}

.policy-text p {
    margin: 0;
    font-size: 14px;
    color: #334155;
}

.back-btn {
    width: 100%;
}

.copyright {
    margin: 4px 0 0;
    text-align: center;
    font-size: 12px;
    color: #64748b;
    letter-spacing: 0.02em;
}

.is-dark .hero-text h2,
.is-dark .info-row strong {
    color: #f8fafc;
}

.is-dark .hero-text p,
.is-dark .info-row span,
.is-dark .about-list,
.is-dark .policy-text p {
    color: #94a3b8;
}

.is-dark .copyright {
    color: #94a3b8;
}

.is-dark .info-row {
    border-color: rgba(148, 163, 184, 0.32);
    background: rgba(15, 23, 42, 0.5);
}
</style>
