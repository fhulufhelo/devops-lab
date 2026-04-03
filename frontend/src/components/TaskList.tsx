import { useEffect, useState } from 'react';
import type { Task } from '../types';
import { getTasks, createTask, updateTask, deleteTask } from '../api';
import TaskForm from './TaskForm';
import TaskItem from './TaskItem';

export default function TaskList() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  async function fetchTasks() {
    try {
      const data = await getTasks();
      setTasks(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tasks');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchTasks();
  }, []);

  async function handleCreate(data: { title: string; description: string }) {
    try {
      const task = await createTask(data);
      setTasks((prev) => [...prev, task]);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create task');
    }
  }

  async function handleStatusChange(id: string, status: Task['status']) {
    try {
      const updated = await updateTask(id, { status });
      setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update task');
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteTask(id);
      setTasks((prev) => prev.filter((t) => t.id !== id));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete task');
    }
  }

  return (
    <div className="task-list">
      <TaskForm onSubmit={handleCreate} />
      {error && <div className="error-banner">{error}</div>}
      {loading ? (
        <p className="loading">Loading tasks…</p>
      ) : tasks.length === 0 ? (
        <p className="empty">No tasks yet. Add one above!</p>
      ) : (
        <div className="tasks">
          {tasks.map((task) => (
            <TaskItem
              key={task.id}
              task={task}
              onStatusChange={handleStatusChange}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
