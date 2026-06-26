<template>
  <section id="sources" class="section sources-section">
    <div class="container">
      <div class="text-center section-header">
        <span class="badge">
          <i class="pi pi-database" />
          Data Sources
        </span>
        <h2 class="section-title">{{ $t("sources.title") }}</h2>
        <p class="section-subtitle text-muted">{{ $t("sources.subtitle") }}</p>
        <div class="divider" />
      </div>

      <div class="sources-grid">
        <div v-for="source in sources" :key="source.key" class="source-card">
          <div class="source-card__header">
            <div class="source-logo" :style="{ background: source.gradient }">
              <i :class="`pi ${source.icon}`" />
            </div>
            <div>
              <h3 class="source-name">{{ $t(`sources.${source.key}.name`) }}</h3>
              <span class="source-tag">{{ source.tag }}</span>
            </div>
          </div>
          <p class="source-desc text-muted">{{ $t(`sources.${source.key}.desc`) }}</p>
          <div class="source-stats">
            <div v-for="stat in source.stats" :key="stat.label" class="source-stat">
              <span class="source-stat__value">{{ stat.value }}</span>
              <span class="source-stat__label">{{ stat.label }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
const sources = [
  {
    key: "circl",
    icon: "pi-shield",
    gradient: "linear-gradient(135deg, #e53e3e, #c53030)",
    tag: "cve.circl.lu",
    stats: [
      { value: "200K+", label: "CVE Records" },
      { value: "CVSS", label: "Scoring" },
      { value: "JSON", label: "Format" },
    ],
  },
  {
    key: "github",
    icon: "pi-github",
    gradient: "linear-gradient(135deg, #742a2a, #9b2c2c)",
    tag: "api.github.com",
    stats: [
      { value: "9.4K+", label: "Advisories" },
      { value: "10+", label: "Ecosystems" },
      { value: "GHSA", label: "IDs" },
    ],
  },
  {
    key: "mycert",
    icon: "pi-flag",
    gradient: "linear-gradient(135deg, #f6ad55, #c05621)",
    tag: "mycert.org.my",
    stats: [
      { value: "790+", label: "Bulletins" },
      { value: "SEA", label: "Region" },
      { value: "HTML", label: "Scraped" },
    ],
  },
];
</script>

<style scoped>
.sources-section {
  background: linear-gradient(180deg, transparent 0%, rgba(99, 102, 241, 0.03) 50%, transparent 100%);
}

.section-header {
  margin-bottom: 56px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.section-title {
  font-size: clamp(1.8rem, 3.5vw, 2.5rem);
  font-weight: 800;
  letter-spacing: -0.02em;
}

.section-subtitle {
  font-size: 1.05rem;
  max-width: 480px;
  line-height: 1.7;
}

.sources-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}

.source-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  transition: border-color 0.2s, transform 0.2s, box-shadow 0.2s;
}

.source-card:hover {
  border-color: rgba(99, 102, 241, 0.3);
  transform: translateY(-2px);
  box-shadow: var(--shadow-glow);
}

.source-card__header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.source-logo {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  color: #fff;
  flex-shrink: 0;
}

.source-name {
  font-size: 1.1rem;
  font-weight: 700;
  margin-bottom: 4px;
}

.source-tag {
  font-size: 0.75rem;
  color: var(--color-text-dim);
  font-family: "SF Mono", monospace;
}

.source-desc {
  font-size: 0.9rem;
  line-height: 1.65;
  flex: 1;
}

.source-stats {
  display: flex;
  gap: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border);
}

.source-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.source-stat__value {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--color-primary-light);
}

.source-stat__label {
  font-size: 0.7rem;
  color: var(--color-text-dim);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

@media (max-width: 1024px) {
  .sources-grid { grid-template-columns: 1fr; }
}
</style>
