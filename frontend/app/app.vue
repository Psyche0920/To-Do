<script setup lang="ts">
// TypeScript describes the task shape; it does not validate JSON at runtime.
interface Task {
  id: number
  title: string
  completed: boolean
}

// Initialization requests GET /api/tasks through the Nuxt proxy.
const taskRequest = await useFetch<Task[]>('/api/tasks', {
  method: 'GET',
  // Use an empty list before data is available.
  default: () => [],
})

// These references expose this request's data and error; they send no new request.
const tasks = taskRequest.data
const error = taskRequest.error
</script>

<template>
  <main>
    <h1>To-Do</h1>

    <p v-if="error" role="alert">
      Unable to load tasks. Check that the Go API is running.
    </p>

    <ul v-else-if="tasks.length > 0">
      <li v-for="task in tasks" :key="task.id">
        {{ task.title }} —
        {{ task.completed ? 'Completed' : 'Not completed' }}
      </li>
    </ul>

    <p v-else>No tasks yet.</p>
  </main>
</template>
