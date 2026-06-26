<template>
  <nav class="nav" :class="{ 'nav--scrolled': scrolled }">
    <div class="container nav__inner">
      <a href="/" class="nav__logo">
        <span class="nav__logo-icon">
          <i class="pi pi-shield" />
        </span>
        <span class="nav__logo-text">NezhaCyberMCP</span>
      </a>

      <ul class="nav__links">
        <li><a href="#features">{{ $t("nav.features") }}</a></li>
        <li><a href="#sources">{{ $t("nav.sources") }}</a></li>
        <li><a href="#tools">{{ $t("nav.tools") }}</a></li>
      </ul>

      <div class="nav__actions">
        <button class="lang-toggle" @click="toggleLang" :title="currentLangLabel">
          <i class="pi pi-globe" />
          <span>{{ currentLangLabel }}</span>
        </button>
        <a
          href="https://github.com/ctkqiang/NezhaCyberMCP"
          target="_blank"
          rel="noopener"
          class="btn-secondary nav__github"
        >
          <i class="pi pi-github" />
          {{ $t("nav.github") }}
        </a>
      </div>

      <button class="nav__hamburger" @click="menuOpen = !menuOpen" :aria-expanded="menuOpen">
        <i :class="menuOpen ? 'pi pi-times' : 'pi pi-bars'" />
      </button>
    </div>

    <div class="nav__mobile" :class="{ 'nav__mobile--open': menuOpen }">
      <a href="#features" @click="menuOpen = false">{{ $t("nav.features") }}</a>
      <a href="#sources" @click="menuOpen = false">{{ $t("nav.sources") }}</a>
      <a href="#tools" @click="menuOpen = false">{{ $t("nav.tools") }}</a>
      <button class="lang-toggle" @click="toggleLang">
        <i class="pi pi-globe" />
        {{ currentLangLabel }}
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
const { locale, setLocale } = useI18n();
const scrolled = ref(false);
const menuOpen = ref(false);

const currentLangLabel = computed(() => (locale.value === "zh" ? "EN" : "中文"));

function toggleLang() {
  setLocale(locale.value === "zh" ? "en" : "zh");
  menuOpen.value = false;
}

onMounted(() => {
  window.addEventListener("scroll", () => {
    scrolled.value = window.scrollY > 20;
  });
});
</script>

<style scoped>
.nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  padding: 16px 0;
  transition: background 0.3s, backdrop-filter 0.3s, border-color 0.3s;
  border-bottom: 1px solid transparent;
}

.nav--scrolled {
  background: rgba(10, 14, 26, 0.85);
  backdrop-filter: blur(16px);
  border-color: var(--color-border);
}

.nav__inner {
  display: flex;
  align-items: center;
  gap: 32px;
}

.nav__logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 1.1rem;
  flex-shrink: 0;
}

.nav__logo-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
}

.nav__links {
  display: flex;
  list-style: none;
  gap: 32px;
  margin-left: auto;
}

.nav__links a {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-muted);
  transition: color 0.2s;
}

.nav__links a:hover {
  color: var(--color-text);
}

.nav__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.lang-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-muted);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
  font-family: var(--font-sans);
}

.lang-toggle:hover {
  border-color: rgba(99, 102, 241, 0.4);
  color: var(--color-text);
}

.nav__github {
  padding: 8px 16px;
  font-size: 0.85rem;
}

.nav__hamburger {
  display: none;
  background: none;
  border: none;
  color: var(--color-text);
  font-size: 1.2rem;
  cursor: pointer;
  margin-left: auto;
}

.nav__mobile {
  display: none;
  flex-direction: column;
  gap: 4px;
  padding: 12px 24px 16px;
  background: rgba(10, 14, 26, 0.95);
  border-top: 1px solid var(--color-border);
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease;
}

.nav__mobile--open {
  max-height: 300px;
}

.nav__mobile a,
.nav__mobile button {
  padding: 12px 0;
  font-size: 1rem;
  color: var(--color-text-muted);
  border-bottom: 1px solid var(--color-border);
  display: block;
  width: 100%;
  text-align: left;
}

@media (max-width: 768px) {
  .nav__links,
  .nav__actions {
    display: none;
  }

  .nav__hamburger {
    display: block;
  }

  .nav__mobile {
    display: flex;
  }
}
</style>
