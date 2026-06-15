import TodoItem from "./TodoItem.jsx";

export default function TodoList({ todos, onDelete, onToggle, onUpdateTitle }) {
  if (todos.length === 0) {
    return (
      <div className="empty-state">
        <h2>No tasks yet</h2>
        <p>Create your first task to start tracking work.</p>
      </div>
    );
  }

  const sortedTodos = [...todos].sort((first, second) => {
    if (first.completed !== second.completed) {
      return first.completed ? 1 : -1;
    }

    return new Date(second.created_at) - new Date(first.created_at);
  });

  return (
    <ul className="todo-list">
      {sortedTodos.map((todo) => (
        <TodoItem
          key={todo.id}
          todo={todo}
          onDelete={onDelete}
          onToggle={onToggle}
          onUpdateTitle={onUpdateTitle}
        />
      ))}
    </ul>
  );
}
