<template>
  <a-card
    class="common-card"
    :bordered="bordered"
    :hoverable="hoverable"
    :style="{ background: background }"
  >
    <template v-if="title || $slots.title" #title>
      <slot name="title">
        <span class="card-title">{{ title }}</span>
      </slot>
    </template>
    <template v-if="$slots.extra" #extra>
      <slot name="extra" />
    </template>
    <slot />
  </a-card>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: ''
  },
  bordered: {
    type: Boolean,
    default: true
  },
  hoverable: {
    type: Boolean,
    default: false
  },
  background: {
    type: String,
    default: ''
  }
})
</script>

<style scoped>
.common-card {
  border-radius: 8px;
  transition: box-shadow 0.3s ease;
}

.common-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-1);
}
</style>
