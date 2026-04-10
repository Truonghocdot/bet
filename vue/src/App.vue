<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AuthLayout from './layouts/AuthLayout.vue'
import MainLayout from './layouts/MainLayout.vue'

const route = useRoute()

const layout = computed(() => (route.meta.layout === 'auth' ? AuthLayout : MainLayout))
</script>

<template>
  <component :is="layout">
    <RouterView v-slot="{ Component, route: childRoute }">
      <Transition name="page" mode="out-in">
        <component :is="Component" :key="childRoute.fullPath" />
      </Transition>
    </RouterView>
  </component>
</template>
