<script setup lang="ts">
// Describe one task returned by the Go API.
interface Task {
  id: number
  title: string
  completed: boolean
}

// Load the task list when the page is initialized.
const taskRequest = await useFetch<Task[]>('/api/tasks', {
  method: 'GET',
  default: () => [],
})

// These references do not send another request.
const tasks = taskRequest.data
const error = taskRequest.error

// Ref values can change while their variable bindings stay constant.
const title = ref('')
const saving = ref(false)
const createError = ref('')

async function addTask(): Promise<void> {
  // Prevent another creation while this submission is in progress.
  if (saving.value) {
    return
  }

  // Clear the previous submission's error.
  createError.value = ''

  const cleanTitle = title.value.trim()
  const titleLength = Array.from(cleanTitle).length

  if (titleLength === 0 || titleLength > 200) {
    createError.value = 'Title must contain 1 to 200 characters.'
    return
  }

  saving.value = true

  try {
    // Wait for creation before clearing input or refreshing the list.
    await $fetch<Task>('/api/tasks', {
      method: 'POST',
      body: { title: cleanTitle },
    })
  } catch {
    createError.value = 'Unable to create task. Please try again.'
    saving.value = false
    return
  }

  title.value = ''

  try {
    // Repeat the GET request to display the saved task.
    await taskRequest.refresh()
  } finally {
    // Re-enable the form when list refreshing finishes.
    saving.value = false
  }
}
</script>

<template>
  <main>
    <h1>To-Do</h1>

    <form @submit.prevent="addTask">
      <label for="task-title">New task</label>

      <input
        id="task-title"
        v-model="title"
        type="text"
        placeholder="What needs to be done?"
        :disabled="saving"
        required
      >

      <button type="submit" :disabled="saving">
        {{ saving ? 'Adding…' : 'Add task' }}
      </button>
    </form>

    <p v-if="createError" role="alert">
      {{ createError }}
    </p>

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