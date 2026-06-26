<template>
  <footer class="footer">
    <div class="container footer__inner">
      <div class="footer__brand">
        <div class="footer__logo">
          <img src="/logo.png" alt="NezhaCyberMCP" class="footer__logo-img" width="28" height="28" />
          <span>NezhaCyberMCP</span>
        </div>
        <p class="footer__tagline text-muted">{{ $t("footer.tagline") }}</p>
      </div>

      <div class="footer__links">
        <a href="https://github.com/ctkqiang/NezhaCyberMCP" target="_blank" rel="noopener">
          <i class="pi pi-github" />
          {{ $t("footer.links.github") }}
        </a>
        <a href="https://github.com/ctkqiang/NezhaCyberMCP/issues" target="_blank" rel="noopener">
          <i class="pi pi-exclamation-circle" />
          {{ $t("footer.links.issues") }}
        </a>
        <a
          href="https://github.com/ctkqiang/NezhaCyberMCP/blob/main/LICENSE"
          target="_blank"
          rel="noopener"
        >
          <i class="pi pi-file" />
          {{ $t("footer.links.license") }}
        </a>
      </div>

      <p class="footer__copy text-muted">{{ $t("footer.copyright") }}</p>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  border-top: 1px solid var(--color-border);
  padding: 48px 0;
}

.footer__inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  text-align: center;
}

.footer__brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.footer__logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 700;
  font-size: 1.1rem;
}

.footer__logo-img {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  object-fit: contain;
  flex-shrink: 0;
}

.footer__tagline {
  font-size: 0.85rem;
}

.footer__links {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
  justify-content: center;
}

.footer__links a {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  color: var(--color-text-muted);
  transition: color 0.2s;
}

.footer__links a:hover {
  color: var(--color-text);
}

.footer__copy {
  font-size: 0.8rem;
}
</style>
