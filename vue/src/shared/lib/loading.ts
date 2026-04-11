import { ref, readonly } from 'vue'

const isLoading = ref(false)

export function useLoading() {
  function setLoading(state: boolean) {
    isLoading.value = state
  }

  return {
    isLoading: readonly(isLoading),
    setLoading
  }
}
