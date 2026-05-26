<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AgencyPanelLayout from '@/layouts/AgencyPanelLayout.vue'
import AgencyPlainLayout from '@/layouts/AgencyPlainLayout.vue'

const route = useRoute()

const layout = computed(() => {
  if (route.meta.layout === 'panel') return AgencyPanelLayout
  return AgencyPlainLayout
})
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>
