import React from "react";
import { useNavigate } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import axios from "axios";

const LikedUsersModal = ({
  show,
  onClose,
  likedUsers,
  title,
  emptytitle,
  onUserClick,
}) => {
  const navigate = useNavigate();

  if (!show) return null;

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
    onClose(); // Close the modal when navigating
  };

  const handleSendPost = async (userid, is_group, postId) => {
    const token = localStorage.getItem("jwtToken");
    if (!token) {
      console.error("JWT token not found");
      return;
    }
    const decoded = jwtDecode(token);
    const loggedInUserId = decoded.sub;

    const payload = {
      sender_id: loggedInUserId,
      receiver_id: userid,
      post_id: Number(postId),
    };

    const endpoint = is_group
      ? `http://localhost:8080/v1/chat/group/${userid}/`
      : `http://localhost:8080/v1/chat/user/${userid}/`;

    try {
      await axios.post(endpoint, payload, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      onClose();
    } catch (error) {
      console.error("Error sending post:", error);
    }
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
            <button
              type="button"
              className="btn-close"
              onClick={onClose}
            ></button>
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
                    onClick={() => {
                      if (onUserClick) {
                        handleSendPost(user.userid, user.is_group, onUserClick);
                      } else {
                        handleUserClick(user.userid);
                      }
                    }}
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
