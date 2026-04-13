<template>
  <Dialog :open="open" @update:open="(val) => emit('update:open', val)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>选择图片来源</DialogTitle>
        <DialogDescription>选择拍照或相册导入。</DialogDescription>
      </DialogHeader>
      <div class="source-grid">
        <Button variant="outline" @click="select('camera')">拍照</Button>
        <Button variant="outline" @click="select('album')">相册</Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'select', value: 'camera' | 'album'): void;
}>();

function select(source: 'camera' | 'album') {
  emit('select', source);
  emit('update:open', false);
}
</script>

<style scoped>
.source-grid {
  display: grid;
  gap: 10px;
}
</style>
