import React, { useState } from "react";
import { useNavigate } from 'react-router-dom';
import axios from "axios";
import LikedUsersModal from "../components/LikedUsersModal";
import ChatList from "../components/ChatList";
import "bootstrap/dist/css/bootstrap.min.css";

const SearchPage = () => {
  const [searchTerm, setSearchTerm] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [showResultsModal, setShowResultsModal] = useState(false);
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [isAddingToGroup, setIsAddingToGroup] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [selectedUsers, setSelectedUsers] = useState([]);
  const navigate = useNavigate();

  const handleKeyDown = async (e, addingToGroup = false) => {
    if (e.key === "Enter") {
      try {
        const token = localStorage.getItem("jwtToken");
        const response = await axios.post(
          "http://localhost:8080/v1/users/search",
          {
            name: searchTerm,
          },
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          }
        );

        setSearchResults(response.data.data || []);
        setShowResultsModal(true);
        setIsAddingToGroup(addingToGroup);
        setSearchTerm("");
      } catch (error) {
        console.error("Error fetching search results:", error);
      }
    }
  };

  const handleShowGroupModalClick = () => {
    setShowGroupModal(true);
  };

  const handleChatSelect = (chat) => {
    if (!chat.is_group) {
      navigate(`/chat/${chat.userid}/${chat.username}`);
    } else {
      navigate(`/group/${chat.userid}/${chat.username}`);
    }
  };

  const handleCreateGroupClick = async () => {
    const token = localStorage.getItem("jwtToken");
    const payload = {
      name: groupName,
      members: selectedUsers.map((user) => user.userid),
    };

    try {
      const response = await axios.post(
        "http://localhost:8080/v1/group/create",
        payload,
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        }
      );

      if (response.status === 200 || response.status === 201) {
        setGroupName("");
        setSelectedUsers([]);
        setShowGroupModal(false);
      } else {
        alert("Failed to create group. Please try again.");
      }
      window.location.reload();
    } catch (error) {
      console.error("Error creating group:", error);
      alert("An error occurred while creating the group.");
    }
  };

  const handleGroupModalClose = () => {
    setShowGroupModal(false);
    setGroupName("");
    setSearchTerm("");
  };

  const handleUserSelect = (user) => {
    setSelectedUsers([...selectedUsers, user]);
    setShowResultsModal(false);
  };

  return (
    <div style={{ height: "100vh", display: "flex", flexDirection: "column" }}>
      {/* Top 10%: Search Box with Create Group Icon */}
      <div
        className="d-flex justify-content-center align-items-center"
        style={{ height: "10vh" }}
      >
        <div
          className="d-flex align-items-center"
          style={{ maxWidth: "500px", width: "100%" }}
        >
          {/* Create Group Icon */}
          <button
            className="btn btn-outline-secondary me-2"
            onClick={handleShowGroupModalClick}
            title="Create Group"
          >
            👥
          </button>

          {/* Search Input */}
          <input
            type="text"
            placeholder="Search users..."
            className="form-control"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={(e) => handleKeyDown(e, false)}
          />
        </div>
      </div>

      {/* Bottom 90%: Reserved for future features */}
      <div style={{ flex: "1", padding: "80px" }}>
        <ChatList onChatSelect={handleChatSelect} />
      </div>

      {/* LikedUsersModal to display search results */}
      <LikedUsersModal
        show={showResultsModal}
        likedUsers={searchResults}
        onClose={() => setShowResultsModal(false)}
        title="Search Results"
        emptytitle="users"
        {...(isAddingToGroup && { onUserSelect: handleUserSelect })}
      />

      {/* Group Creation Modal */}
      {showGroupModal && (
        <div className="modal fade show d-block" tabIndex="-1" role="dialog">
          <div className="modal-dialog" role="document">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Create New Group</h5>
                <button
                  type="button"
                  className="btn-close"
                  onClick={handleGroupModalClose}
                  aria-label="Close"
                ></button>
              </div>
              <div className="modal-body">
                <div className="mb-3">
                  <label htmlFor="groupName" className="form-label">
                    Group Name
                  </label>
                  <input
                    type="text"
                    className="form-control"
                    id="groupName"
                    placeholder="Enter group name"
                    value={groupName}
                    onChange={(e) => setGroupName(e.target.value)}
                  />
                </div>
                <div className="mb-3">
                  <label htmlFor="memberSearch" className="form-label">
                    Add Members
                  </label>
                  <input
                    type="text"
                    className="form-control"
                    id="memberSearch"
                    placeholder="Search and add members"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onKeyDown={(e) => handleKeyDown(e, true)}
                  />
                </div>
                {selectedUsers.length > 0 && (
                  <div className="mt-3">
                    <div className="d-flex flex-wrap">
                      {selectedUsers.map((user) => (
                        <span
                          key={user.userid}
                          className="badge bg-primary me-2 mb-2"
                        >
                          {user.username}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
              <div className="modal-footer">
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={handleGroupModalClose}
                >
                  Close
                </button>
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={handleCreateGroupClick}
                >
                  Create Group
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SearchPage;
