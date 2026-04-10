<template>
    <section class="app-page app-inner-page pt-8 about-page" :class="{ 'is-dark': resolvedTheme === 'dark' }">
        <header class="app-page-header page-header">
            <Button variant="outline" size="icon-sm" class="app-header-back mr-4" @click="goBack" aria-label="返回上一级">< </Button>
            <span class="app-kicker">About</span>
            <h1>关于应用</h1>
            <p>了解产品定位、版本信息和数据处理方式。</p>
        </header>

        <Card class="app-page-shell about-hero">
            <CardContent class="hero-content">
                <div class="hero-mark">LQ</div>
                <div class="hero-text">
                    <h2>{{ appName }}</h2>
                    <p>为 LaTeX 错题整理、回顾与强化练习而设计。</p>
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
                <CardDescription>产品定位</CardDescription>
                <CardTitle>我们在做什么</CardTitle>
            </CardHeader>
            <CardContent>
                <ul class="about-list">
                    <li>让错题录入、结构化和复习闭环更快完成。</li>
                    <li>把题目、答案、错因统一成可追踪的学习记录。</li>
                    <li>在移动端保持清晰、轻量、可回溯的使用体验。</li>
                </ul>
            </CardContent>
        </Card>

        <Card class="app-soft-card info-card">
            <CardHeader>
                <CardDescription>隐私说明</CardDescription>
                <CardTitle>数据与权限</CardTitle>
            </CardHeader>
            <CardContent class="policy-text">
                <p>仅在题目处理流程中使用你主动选择的图片内容。</p>
                <p>账号与学习记录用于提供同步与统计能力，不用于广告投放。</p>
            </CardContent>
        </Card>

        <Button variant="outline" class="back-btn" @click="goBack">返回我的页面</Button>
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
    background: linear-gradient(145deg, #22c55e 0%, #16a34a 50%, #0f766e 100%);
    box-shadow: 0 10px 24px rgba(22, 163, 74, 0.28);
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

.is-dark .info-row {
    border-color: rgba(148, 163, 184, 0.32);
    background: rgba(15, 23, 42, 0.5);
}
</style>
