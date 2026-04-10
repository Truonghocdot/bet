<template>
  <div class="app-shell">
    <div class="app-shell__decor">
      <div class="app-shell__blur app-shell__blur--one"></div>
      <div class="app-shell__blur app-shell__blur--two"></div>
    </div>

    <main class="app-main">
      <div class="mx-auto min-h-dvh w-full max-w-[560px] px-5 pb-8 pt-3 md:px-6 md:pb-10 md:pt-8">
        <slot />
      </div>
    </main>
  </div>
</template>
