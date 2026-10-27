<script setup lang="ts">
// Describe one task returned by the Go API.
interface Task {
  id: number
  title: string
  completed: boolean
}

// Load tasks when the page is initialized.
const taskRequest = await useFetch<Task[]>('/api/tasks', {
  method: 'GET',
  default: () => [],
})

const tasks = taskRequest.data
const error = taskRequest.error

// Creation state: ref values are accessed with .value in scripts.
const title = ref('')
const saving = ref(false)
const createError = ref('')

// Completion-update state is separate from creation state.
// Busy only while saving a completion change.
const toggleTaskUpdating = ref(false)
const toggleTaskError = ref('')

async function addTask(): Promise<void> {
  // Ignore another creation while one is in progress.
  if (saving.value) {
    return
  }

  createError.value = ''

  const cleanTitle = title.value.trim()
  const titleLength = Array.from(cleanTitle).length

  if (titleLength === 0 || titleLength > 200) {
    createError.value = 'Title must contain 1 to 200 characters.'
    return
  }

  saving.value = true

  try {
    // Wait for the server to save the task.
    await $fetch<Task>('/api/tasks', {
      method: 'POST',
      body: { title: cleanTitle },
    })
  } catch {
    createError.value = 'Unable to create task. Please try again.'
    saving.value = false
    return
  }

  // Clear the reusable input, not the saved database task.
  title.value = ''

  try {
    await taskRequest.refresh()
  } finally {
    saving.value = false
  }
}

// Promise<void> represents completion without returning a task to the caller.
async function toggleTask(task: Task): Promise<void> {
  // task is the selected item; Task is its type.
  if (toggleTaskUpdating.value || deleting.value) {
    return
  }

  toggleTaskError.value = ''
  toggleTaskUpdating.value = true

  try {
    // ! reverses the current completion state.
    await $fetch<Task>(`/api/tasks/${task.id}`, {
      method: 'PATCH',
      body: { completed: !task.completed },
    })
  } catch {
    toggleTaskError.value = 'Unable to update task. Please try again.'
    toggleTaskUpdating.value = false

    // Restore the checkbox display from the stored task state.
    await taskRequest.refresh()
    return
  }

  try {
    // Repeat the list request after a successful update.
    await taskRequest.refresh()
  } finally {
    toggleTaskUpdating.value = false
  }
}

// Deletion has its own busy state and message.
const deleting = ref(false)
const deleteTaskError = ref('')

async function deleteTask(task: Task): Promise<void> {
  // Avoid overlapping deletion and completion updates.
  if (deleting.value || toggleTaskUpdating.value) {
    return
  }

  deleteTaskError.value = ''
  deleting.value = true

  try {
    // DELETE returns no task body on success.
    await $fetch(`/api/tasks/${task.id}`, {
      method: 'DELETE',
    })
  } catch {
    deleteTaskError.value = 'Unable to delete task. Please try again.'
    deleting.value = false
    return
  }

  try {
    // Reload saved tasks so the deleted item disappears.
    await taskRequest.refresh()
  } finally {
    deleting.value = false
  }
}

</script>

<template>
  <main>
    <h1>Wehan To-Do List</h1>

    <form @submit.prevent="addTask">
      <label for="task-title"></label>

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

    <!-- Error messages are independent display rules, not function calls. -->
    <p v-if="createError" role="alert">
      {{ createError }}
    </p>

    <p v-if="toggleTaskError" role="alert">
      {{ toggleTaskError }}
    </p>

    <p v-if="deleteTaskError" role="alert">
      {{ deleteTaskError }}
    </p>

    <!-- This if/else-if/else group selects one list display state. -->
    <p v-if="error" role="alert">
      Unable to load tasks. Check that the Go API is running.
    </p>

    <ul v-else-if="tasks.length > 0">
      <li v-for="task in tasks" :key="task.id">
        <label>
          <!-- checked displays stored state; change triggers the script function. -->
          <input
            type="checkbox"
            :checked="task.completed"
            :disabled="toggleTaskUpdating || deleting"
            @change="toggleTask(task)"
          >
          {{ task.title }}
        </label>

        — {{ task.completed ? 'Completed' : 'Not completed' }}

        <!-- click triggers deletion; aria-label identifies the task for screen readers. -->
        <button
          type="button"
          :disabled="deleting || toggleTaskUpdating"
          :aria-label="`Delete ${task.title}`"
          @click="deleteTask(task)"
        >
          Delete
        </button>
      </li>
    </ul>

    <p v-else>No tasks yet.</p>
  </main>
</template>
