import React from "react";
import { useNavigate } from "react-router-dom";

const LikedUsersModal = ({ show, onClose, likedUsers, title, emptytitle }) => {
  const navigate = useNavigate();

  if (!show) return null;

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
    onClose(); // Close the modal when navigating
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
            <h5 className="modal-title">{title}</h5>
            <button type="button" className="btn-close" onClick={onClose}></button>
          </div>
          <div className="modal-body">
            {likedUsers.length === 0 ? (
              <p>No {emptytitle} yet.</p>
            ) : (
              <ul className="list-group">
                {likedUsers.map((user) => (
                  <li
                    key={user.userid}
                    className="list-group-item"
                    style={{ cursor: "pointer", color: "#0d6efd" }}
                    onClick={() => handleUserClick(user.userid)}
                  >
                    @{user.username}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default LikedUsersModal;
