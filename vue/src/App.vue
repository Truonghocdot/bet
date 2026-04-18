<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AuthLayout from './layouts/AuthLayout.vue'
import AdminLayout from './layouts/AdminLayout.vue'
import MainLayout from './layouts/MainLayout.vue'

const route = useRoute()

const layout = computed(() => {
  if (route.meta.layout === 'auth') return AuthLayout
  if (route.meta.layout === 'admin') return AdminLayout
  return MainLayout
})
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
