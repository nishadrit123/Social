import React from "react";
import { useNavigate } from "react-router-dom";

const CommentsModal = ({
  show,
  onClose,
  comments,
  newComment,
  setNewComment,
  handlePostComment,
  editingCommentId,
  editText,
  setEditText,
  handleSaveEdit,
  handleCancelEdit,
  handleEditClick,
  handleDeleteComment,
}) => {
  const navigate = useNavigate();

  if (!show) return null;

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
    onClose();
  };

  return (
    <div className="modal d-block" tabIndex="-1" onClick={onClose}>
      <div
        className="modal-dialog modal-dialog-centered"
        role="document"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">Comments</h5>
            <button type="button" className="btn-close" onClick={onClose}></button>
          </div>
          <div className="modal-body">
            {comments.length === 0 ? (
              <p className="text-muted">No comments yet.</p>
            ) : (
              <ul className="list-group">
                {comments.map((commentObj, index) => (
                  <li
                    key={index}
                    className="list-group-item d-flex justify-content-between align-items-center"
                  >
                    <div className="w-100">
                      <strong
                        style={{ cursor: "pointer", color: "#0d6efd" }}
                        onClick={() => handleUserClick(commentObj.userid)}
                      >
                        @{commentObj.username}:
                      </strong>{" "}
                      {editingCommentId === commentObj.id ? (
                        <div>
                          <input
                            type="text"
                            className="form-control mt-1"
                            value={editText}
                            onChange={(e) => setEditText(e.target.value)}
                          />
                          <div className="mt-1 d-flex gap-2">
                            <button
                              className="btn btn-sm btn-success"
                              onClick={() => handleSaveEdit(commentObj)}
                            >
                              ✅ Save
                            </button>
                            <button
                              className="btn btn-sm btn-secondary"
                              onClick={handleCancelEdit}
                            >
                              ❌ Cancel
                            </button>
                          </div>
                        </div>
                      ) : (
                        <span>{commentObj.comment}</span>
                      )}
                    </div>

                    {commentObj.should_update_delete && (
                      <div className="ms-2 mt-1">
                        <span
                          className="me-2"
                          style={{ cursor: "pointer" }}
                          onClick={() => handleEditClick(commentObj)}
                        >
                          ✏️
                        </span>
                        <span
                          style={{ cursor: "pointer" }}
                          onClick={() => handleDeleteComment(commentObj)}
                        >
                          🗑️
                        </span>
                      </div>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
          <div className="modal-footer">
            <input
              type="text"
              className="form-control"
              placeholder="Add a comment..."
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
            />
            <button
              className="btn btn-primary"
              onClick={handlePostComment}
              disabled={!newComment.trim()}
            >
              Comment
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CommentsModal;
