<template>
  <div class="admin-shell">
    <div class="admin-shell__decor" aria-hidden="true">
      <div class="admin-shell__blur admin-shell__blur--one"></div>
      <div class="admin-shell__blur admin-shell__blur--two"></div>
    </div>

    <main class="admin-shell__main">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
  background:
    radial-gradient(circle at 14% 12%, rgba(34, 197, 94, 0.16), transparent 0 28%),
    radial-gradient(circle at 84% 10%, rgba(14, 165, 233, 0.18), transparent 0 32%),
    linear-gradient(180deg, #040816 0%, #081224 52%, #07101f 100%);
  position: relative;
  overflow: hidden;
}

.admin-shell__decor {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.admin-shell__blur {
  position: absolute;
  border-radius: 999px;
  filter: blur(80px);
  opacity: 0.28;
}

.admin-shell__blur--one {
  width: 320px;
  height: 320px;
  top: -100px;
  left: -80px;
  background: #22c55e;
}

.admin-shell__blur--two {
  width: 360px;
  height: 360px;
  top: -140px;
  right: -110px;
  background: #0ea5e9;
}

.admin-shell__main {
  position: relative;
  z-index: 1;
}
</style>
