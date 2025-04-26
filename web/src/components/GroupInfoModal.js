import React from 'react';
import '../style/GroupInfoModal.css';
import { useNavigate } from "react-router-dom";

const GroupInfoModal = ({ groupInfo, onClose }) => {
  const navigate = useNavigate();

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <button className="close-button" onClick={onClose}>×</button>
        <h2>{groupInfo.name}</h2>
        <p style={{ cursor: "pointer", color: "#0d6efd", display: "block" }}
        onClick={() => handleUserClick(groupInfo.admin.id)}>
          <h3>👤</h3> 
          {groupInfo.admin.username}
        </p>
        <h3>👥</h3>
        <ul>
          {groupInfo.members.map((member) => (
            <li key={member.id} style={{ cursor: "pointer", color: "#0d6efd", display: "block" }}
            onClick={() => handleUserClick(member.id)}>
              {member.username}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default GroupInfoModal;
