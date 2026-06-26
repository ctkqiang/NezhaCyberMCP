<template>
  <section class="hero">
    <div class="hero__bg">
      <div class="hero__orb hero__orb--1" />
      <div class="hero__orb hero__orb--2" />
      <div class="hero__grid" />
    </div>

    <div class="container hero__inner">
      <div class="hero__content animate-fade-up">
        <span class="badge">
          <i class="pi pi-server" />
          {{ $t("hero.badge") }}
        </span>

        <h1 class="hero__title">
          <span class="gradient-text">{{ $t("hero.subtitle") }}</span>
          <br />
          {{ $t("hero.title") }}
        </h1>

        <p class="hero__desc">{{ $t("hero.description") }}</p>

        <div class="hero__actions">
          <a href="#tools" class="btn-primary">
            <i class="pi pi-play" />
            {{ $t("hero.cta_primary") }}
          </a>
          <a
            href="https://github.com/ctkqiang/NezhaCyberMCP"
            target="_blank"
            rel="noopener"
            class="btn-secondary"
          >
            <i class="pi pi-github" />
            {{ $t("hero.cta_secondary") }}
          </a>
        </div>

        <div class="hero__stats">
          <div v-for="stat in stats" :key="stat.key" class="hero__stat">
            <span class="hero__stat-value">{{ stat.value }}</span>
            <span class="hero__stat-label">{{ $t(`hero.stats.${stat.key}`) }}</span>
          </div>
        </div>
      </div>

      <div class="hero__visual animate-float">
        <div class="terminal">
          <div class="terminal__bar">
            <span class="terminal__dot terminal__dot--red" />
            <span class="terminal__dot terminal__dot--yellow" />
            <span class="terminal__dot terminal__dot--green" />
            <span class="terminal__title">MCP Inspector</span>
          </div>
          <div class="terminal__body">
            <div class="terminal__line">
              <span class="terminal__prompt">$</span>
              <span class="terminal__cmd"> tools/call get_cve</span>
            </div>
            <div class="terminal__line terminal__line--param">
              <span class="terminal__key">  cve_id:</span>
              <span class="terminal__val"> "CVE-2021-44228"</span>
            </div>
            <div class="terminal__line terminal__line--divider">─────────────────────────</div>
            <div class="terminal__line terminal__line--result">
              <span class="terminal__key">  title:</span>
              <span class="terminal__val"> "Apache Log4j2 RCE"</span>
            </div>
            <div class="terminal__line terminal__line--result">
              <span class="terminal__key">  severity:</span>
              <span class="terminal__badge terminal__badge--critical"> CRITICAL</span>
            </div>
            <div class="terminal__line terminal__line--result">
              <span class="terminal__key">  cvss:</span>
              <span class="terminal__val"> 10.0</span>
            </div>
            <div class="terminal__line terminal__line--result">
              <span class="terminal__key">  source:</span>
              <span class="terminal__val"> "circl_cves"</span>
            </div>
            <div class="terminal__line terminal__line--result">
              <span class="terminal__key">  cwe:</span>
              <span class="terminal__val"> ["CWE-502"]</span>
            </div>
            <div class="terminal__cursor" />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
const stats = [
  { key: "cves", value: "200K+" },
  { key: "advisories", value: "9.4K+" },
  { key: "mycert", value: "790+" },
  { key: "tools", value: "10" },
];
</script>

<style scoped>
.hero {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  padding: 120px 0 80px;
  overflow: hidden;
}

.hero__bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.hero__orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.15;
}

.hero__orb--1 {
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, #e53e3e 0%, #9b2c2c 100%);
  top: -200px;
  right: -100px;
}

.hero__orb--2 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, #f6ad55 0%, #c05621 100%);
  bottom: -100px;
  left: -100px;
}

.hero__grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
  background-size: 60px 60px;
}

.hero__inner {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  align-items: center;
  position: relative;
  z-index: 1;
}

.hero__content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.hero__title {
  font-size: clamp(2.2rem, 5vw, 3.5rem);
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.02em;
}

.hero__desc {
  font-size: 1.1rem;
  color: var(--color-text-muted);
  line-height: 1.7;
  max-width: 520px;
}

.hero__actions {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.hero__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding-top: 8px;
  border-top: 1px solid var(--color-border);
}

.hero__stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.hero__stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--color-primary-light), var(--color-accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero__stat-label {
  font-size: 0.75rem;
  color: var(--color-text-dim);
  font-weight: 500;
}

.terminal {
  background: var(--terminal-bg);
  border: 1px solid var(--terminal-border);
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.35), 0 0 0 1px rgba(79, 110, 247, 0.1);
  font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
  font-size: 0.85rem;
}

.terminal__bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--terminal-bar);
  border-bottom: 1px solid var(--terminal-border);
}

.terminal__dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.terminal__dot--red { background: #ff5f57; }
.terminal__dot--yellow { background: #febc2e; }
.terminal__dot--green { background: #28c840; }

.terminal__title {
  margin-left: auto;
  font-size: 0.75rem;
  color: var(--terminal-dim);
}

.terminal__body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.terminal__line {
  display: flex;
  align-items: center;
}

.terminal__prompt {
  color: #28c840;
  font-weight: 700;
}

.terminal__cmd {
  color: #e6edf3;
}

.terminal__line--param .terminal__key {
  color: #79c0ff;
}

.terminal__line--param .terminal__val {
  color: #a5d6ff;
}

.terminal__line--divider {
  color: #30363d;
  font-size: 0.75rem;
  margin: 4px 0;
}

.terminal__line--result .terminal__key {
  color: #7ee787;
}

.terminal__line--result .terminal__val {
  color: #e6edf3;
}

.terminal__badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
}

.terminal__badge--critical {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.terminal__cursor {
  width: 8px;
  height: 16px;
  background: var(--color-primary-light);
  border-radius: 2px;
  animation: blink 1s step-end infinite;
  margin-top: 4px;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

@media (max-width: 1024px) {
  .hero__inner {
    grid-template-columns: 1fr;
    gap: 48px;
  }

  .hero__stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .hero__visual {
    max-width: 560px;
    margin: 0 auto;
    width: 100%;
  }
}

@media (max-width: 768px) {
  .hero {
    padding: 100px 0 64px;
    min-height: auto;
  }

  .hero__title {
    font-size: clamp(1.8rem, 7vw, 2.6rem);
  }

  .hero__desc {
    font-size: 1rem;
  }

  .hero__actions {
    flex-direction: column;
    gap: 12px;
  }

  .hero__stats {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .hero__stat-value {
    font-size: 1.25rem;
  }

  /* Terminal — prevent horizontal overflow */
  .terminal {
    font-size: 0.78rem;
  }

  .terminal__body {
    padding: 14px;
  }

  .terminal__line--divider {
    display: none;
  }
}

@media (max-width: 480px) {
  .hero {
    padding: 88px 0 48px;
  }

  .hero__stats {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .hero__actions {
    flex-direction: column;
  }

  .terminal {
    font-size: 0.72rem;
    border-radius: 12px;
  }
}
</style>
