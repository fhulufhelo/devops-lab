import type { Task } from '../types';

const STATUS_ORDER: Task['status'][] = ['todo', 'in_progress', 'done'];

const STATUS_LABELS: Record<Task['status'], string> = {
  todo: 'To Do',
  in_progress: 'In Progress',
  done: 'Done',
};

interface TaskItemProps {
  task: Task;
  onStatusChange: (id: string, status: Task['status']) => void;
  onDelete: (id: string) => void;
}

export default function TaskItem({ task, onStatusChange, onDelete }: TaskItemProps) {
  function cycleStatus() {
    const currentIndex = STATUS_ORDER.indexOf(task.status);
    const nextStatus = STATUS_ORDER[(currentIndex + 1) % STATUS_ORDER.length];
    onStatusChange(task.id, nextStatus);
  }

  return (
    <div className="task-item">
      <div className="task-content">
        <div className="task-header">
          <span className="task-title">{task.title}</span>
          <span className={`status-badge status-${task.status}`}>
            {STATUS_LABELS[task.status]}
          </span>
        </div>
        {task.description && <p className="task-description">{task.description}</p>}
      </div>
      <div className="task-actions">
        <button className="btn-status" onClick={cycleStatus}>Next Status</button>
        <button className="btn-delete" onClick={() => onDelete(task.id)}>Delete</button>
      </div>
    </div>
  );
}
