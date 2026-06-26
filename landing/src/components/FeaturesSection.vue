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
  { key: "realtime", icon: "pi-sync",     gradient: "linear-gradient(135deg, #e53e3e, #c53030)" },
  { key: "search",   icon: "pi-search",   gradient: "linear-gradient(135deg, #f6ad55, #c05621)" },
  { key: "lambda",   icon: "pi-cloud",    gradient: "linear-gradient(135deg, #fc8181, #e53e3e)" },
  { key: "secure",   icon: "pi-lock",     gradient: "linear-gradient(135deg, #9b2c2c, #742a2a)" },
  { key: "prompts",  icon: "pi-comment",  gradient: "linear-gradient(135deg, #e53e3e, #f6ad55)" },
  { key: "db",       icon: "pi-database", gradient: "linear-gradient(135deg, #c53030, #9b2c2c)" },
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
