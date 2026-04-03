import './App.css';
import TaskList from './components/TaskList';

export default function App() {
  return (
    <div className="app">
      <header className="app-header">
        <h1>Task Tracker</h1>
      </header>
      <main>
        <TaskList />
      </main>
    </div>
  );
}
