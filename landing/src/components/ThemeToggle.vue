<template>
  <button
    class="theme-toggle"
    :class="{ 'theme-toggle--light': !isDark }"
    :aria-label="isDark ? '切换到浅色模式' : '切换到深色模式'"
    :title="isDark ? '切换到浅色模式 / Switch to Light Mode' : '切换到深色模式 / Switch to Dark Mode'"
    @click="toggle"
  >
    <span class="theme-toggle__track">
      <span class="theme-toggle__thumb">
        <i :class="isDark ? 'pi pi-moon' : 'pi pi-sun'" />
      </span>
    </span>
  </button>
</template>

<script setup lang="ts">
const { isDark, toggle } = useTheme();
</script>

<style scoped>
.theme-toggle {
  display: inline-flex;
  align-items: center;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  outline: none;
}

.theme-toggle:focus-visible .theme-toggle__track {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.theme-toggle__track {
  position: relative;
  width: 48px;
  height: 26px;
  border-radius: 999px;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border-2);
  transition: background 0.25s ease, border-color 0.25s ease;
  display: flex;
  align-items: center;
}

.theme-toggle--light .theme-toggle__track {
  background: #dbeafe;
  border-color: rgba(58, 86, 224, 0.25);
}

.theme-toggle__thumb {
  position: absolute;
  left: 3px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e53e3e, #c53030);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  color: #fff;
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.theme-toggle--light .theme-toggle__thumb {
  transform: translateX(22px);
  background: linear-gradient(135deg, #f6ad55, #c05621);
}
</style>
