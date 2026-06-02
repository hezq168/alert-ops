<template>
  <!-- 如果有子菜单，渲染为 el-sub-menu -->
  <el-sub-menu v-if="menu.children && menu.children.length > 0" :index="String(menu.id)">
    <template #title>
      <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
      <span>{{ menu.name }}</span>
    </template>
    <menu-item v-for="child in menu.children" :key="child.id" :menu="child" />
  </el-sub-menu>

  <!-- 如果没有子菜单，渲染为 el-menu-item -->
  <el-menu-item v-else :index="menu.path" v-if="!menu.hidden">
    <el-icon v-if="menu.icon"><component :is="menu.icon" /></el-icon>
    <template #title>{{ menu.name }}</template>
  </el-menu-item>
</template>

<script setup>
defineProps({
  menu: {
    type: Object,
    required: true
  }
})
</script>