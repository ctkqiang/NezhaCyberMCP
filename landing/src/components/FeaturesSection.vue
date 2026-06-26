<template>
  <section id="features" class="section">
    <div class="container">
      <div class="text-center section-header">
        <span class="badge">
          <i class="pi pi-star" />
          Features
        </span>
        <h2 class="section-title">{{ $t("features.title") }}</h2>
        <p class="section-subtitle text-muted">{{ $t("features.subtitle") }}</p>
        <div class="divider" />
      </div>

      <div class="grid-3 features-grid">
        <div v-for="feature in features" :key="feature.key" class="card feature-card">
          <div class="feature-icon" :style="{ background: feature.gradient }">
            <i :class="`pi ${feature.icon}`" />
          </div>
          <h3 class="feature-title">{{ $t(`features.items.${feature.key}.title`) }}</h3>
          <p class="feature-desc text-muted">{{ $t(`features.items.${feature.key}.desc`) }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
const features = [
  { key: "realtime", icon: "pi-sync", gradient: "linear-gradient(135deg, #6366f1, #818cf8)" },
  { key: "search", icon: "pi-search", gradient: "linear-gradient(135deg, #06b6d4, #22d3ee)" },
  { key: "lambda", icon: "pi-cloud", gradient: "linear-gradient(135deg, #f59e0b, #fbbf24)" },
  { key: "secure", icon: "pi-lock", gradient: "linear-gradient(135deg, #10b981, #34d399)" },
  { key: "prompts", icon: "pi-comment", gradient: "linear-gradient(135deg, #8b5cf6, #a78bfa)" },
  { key: "db", icon: "pi-database", gradient: "linear-gradient(135deg, #ef4444, #f87171)" },
];
</script>

<style scoped>
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
  max-width: 640px;
}

.section-subtitle {
  font-size: 1.05rem;
  max-width: 560px;
  line-height: 1.7;
}

.features-grid {
  margin-top: 0;
}

.feature-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-icon {
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

.feature-title {
  font-size: 1.05rem;
  font-weight: 700;
}

.feature-desc {
  font-size: 0.9rem;
  line-height: 1.65;
}
</style>
