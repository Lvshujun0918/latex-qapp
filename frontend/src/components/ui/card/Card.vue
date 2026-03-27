<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    :class="
      cn(
        'rounded-xl border bg-card text-card-foreground shadow transition-colors dark:border-slate-700/60 dark:bg-slate-900/65',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
