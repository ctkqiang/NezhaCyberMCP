<template>
  <nav class="nav" :class="{ 'nav--scrolled': scrolled }">
    <div class="container nav__inner">
      <a href="/" class="nav__logo">
        <img src="/logo.png" alt="NezhaCyberMCP" class="nav__logo-img" width="32" height="32" />
        <span class="nav__logo-text">NezhaCyberMCP</span>
      </a>

      <ul class="nav__links">
        <li><a href="#features">{{ $t("nav.features") }}</a></li>
        <li><a href="#sources">{{ $t("nav.sources") }}</a></li>
        <li><a href="#tools">{{ $t("nav.tools") }}</a></li>
      </ul>

      <div class="nav__actions">
        <ThemeToggle />
        <button class="lang-toggle" @click="toggleLang" :title="switchLabel">
          <span class="lang-toggle__current">{{ currentFlag }} {{ currentLangName }}</span>
          <span class="lang-toggle__sep">/</span>
          <span class="lang-toggle__switch">{{ switchLabel }}</span>
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
        <span>{{ currentFlag }} {{ currentLangName }}</span>
        <span class="lang-toggle__sep">/</span>
        <span class="lang-toggle__switch">{{ switchLabel }}</span>
      </button>
      <div class="nav__mobile-theme">
        <ThemeToggle />
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
const { locale, setLocale } = useI18n();
const scrolled = ref(false);
const menuOpen = ref(false);

const currentLangLabel = computed(() => (locale.value === "zh" ? "EN" : "中文"));
const currentFlag     = computed(() => (locale.value === "zh" ? "🇨🇳" : "🇺🇸"));
const currentLangName = computed(() => (locale.value === "zh" ? "中文" : "EN"));
const switchLabel     = computed(() => (locale.value === "zh" ? "EN" : "中文"));

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
  background: var(--nav-bg-scrolled);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-color: var(--color-border);
  box-shadow: var(--shadow-nav);
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

.nav__logo-img {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  object-fit: contain;
  flex-shrink: 0;
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
  gap: 5px;
  padding: 7px 12px;
  border-radius: 8px;
  background: transparent;
  border: 1px solid var(--color-border-2);
  color: var(--color-text-muted);
  font-size: 0.82rem;
  font-weight: 500;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s, background 0.2s;
  font-family: "Inter", "Noto Sans SC", system-ui, sans-serif;
  white-space: nowrap;
}

.lang-toggle:hover {
  border-color: rgba(229, 62, 62, 0.45);
  color: var(--color-text);
  background: rgba(229, 62, 62, 0.06);
}

.lang-toggle__current {
  color: var(--color-text);
  font-weight: 600;
}

.lang-toggle__sep {
  color: var(--color-text-dim);
  font-size: 0.75rem;
}

.lang-toggle__switch {
  color: var(--color-primary-light);
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

.nav__mobile-theme {
  padding: 12px 0;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text-muted);
  font-size: 0.9rem;
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
