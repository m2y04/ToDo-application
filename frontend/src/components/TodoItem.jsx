import { useState } from "react";

export default function TodoItem({ todo, onDelete, onToggle, onUpdateTitle }) {
  const [isEditing, setIsEditing] = useState(false);
  const [draftTitle, setDraftTitle] = useState(todo.title);

  async function handleSave(event) {
    event.preventDefault();

    const trimmedTitle = draftTitle.trim();
    if (!trimmedTitle || trimmedTitle === todo.title) {
      setDraftTitle(todo.title);
      setIsEditing(false);
      return;
    }

    await onUpdateTitle(todo, trimmedTitle);
    setIsEditing(false);
  }

  function handleCancel() {
    setDraftTitle(todo.title);
    setIsEditing(false);
  }

  return (
    <li className={`todo-item ${todo.completed ? "todo-item--completed" : ""}`}>
      <input
        aria-label={`Mark ${todo.title} as ${todo.completed ? "active" : "complete"}`}
        checked={todo.completed}
        className="todo-checkbox"
        type="checkbox"
        onChange={() => onToggle(todo)}
      />

      <div className="todo-content">
        {isEditing ? (
          <form className="edit-form" onSubmit={handleSave}>
            <input
              aria-label="Edit todo title"
              autoFocus
              value={draftTitle}
              onChange={(event) => setDraftTitle(event.target.value)}
            />
            <div className="row-actions">
              <button className="button button--small" disabled={!draftTitle.trim()} type="submit">
                Save
              </button>
              <button className="button button--ghost button--small" type="button" onClick={handleCancel}>
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <>
            <p className="todo-title">{todo.title}</p>
            <p className="todo-meta">
              Updated {new Date(todo.updated_at || todo.created_at).toLocaleString()}
            </p>
          </>
        )}
      </div>

      {!isEditing ? (
        <div className="row-actions">
          <button className="button button--ghost button--small" type="button" onClick={() => setIsEditing(true)}>
            Edit
          </button>
          <button className="button button--danger button--small" type="button" onClick={() => onDelete(todo)}>
            Delete
          </button>
        </div>
      ) : null}
    </li>
  );
}
