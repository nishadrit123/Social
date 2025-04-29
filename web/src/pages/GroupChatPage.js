import React, { useEffect, useState, useRef } from "react";
import { useParams } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import axios from "axios";
import "../style/ChatPage.css";
import PostCard from "../components/PostCard";
import GroupInfoModal from "../components/GroupInfoModal";
import LikedUsersModal from "../components/LikedUsersModal";
import { useNavigate } from "react-router-dom";

const GroupChatPage = () => {
  const { userId, username } = useParams();
  const [messages, setMessages] = useState([]);
  const [messageText, setMessageText] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [groupInfo, setGroupInfo] = useState(null);
  const [showGroupInfo, setShowGroupInfo] = useState(false);
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [showResultsModal, setShowResultsModal] = useState(false);
  const [searchResults, setSearchResults] = useState([]);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const chatMessagesRef = useRef(null);
  const socketRef = useRef(null);

  const token = localStorage.getItem("jwtToken");
  const decoded = jwtDecode(token);
  const loggedInUserId = decoded.sub;
  const navigate = useNavigate();

  useEffect(() => {
    const fetchGroupChat = async () => {
      try {
        const response = await axios.get(
          `http://localhost:8080/v1/chat/group/${userId}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
        setMessages(response.data.data || []);
      } catch (error) {
        console.error("Error fetching group chat:", error);
      }
    };

    fetchGroupChat();
  }, [userId, token]);   

  // useEffect(() => {
  //   const grpMembers = async () => {
  //     try {
  //       const response = await axios.get(
  //         `http://localhost:8080/v1/group/info/${userId}`,
  //         {
  //           headers: {
  //             Authorization: `Bearer ${token}`,
  //           },
  //         }
  //       );
  //       const info = response.data.data;
  //       if (loggedInUserId !== info.admin.id) {
  //         members.push(info.admin.id);
  //       }
  //       info.members.forEach((member) => {
  //         if (loggedInUserId !== member.id) {
  //           members.push(member.id);
  //         }
  //       });
  //       const unique = new Set(members)
  //       unique_members = [...unique] 
  //       console.log(unique_members);
  //     } catch (error) {
  //       console.error("Error fetching group info:", error);
  //     }
  //   };
  //   grpMembers();
  // }, [userId, token]);

  useEffect(() => {
    let unique_members = [];
    const members = [];
    const prepare = async() => {
    const response = await axios.get(
      `http://localhost:8080/v1/group/info/${userId}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
    const info = response.data.data;
    if (loggedInUserId !== info.admin.id) {
      members.push(info.admin.id);
    }
    info.members.forEach((member) => {
      if (loggedInUserId !== member.id) {
        members.push(member.id);
      }
    });
    const unique = new Set(members)
    unique_members = [...unique] 
    unique_members.unshift(userId) // prepend groupID to array for websocket display
    socketRef.current = new WebSocket(`ws://localhost:5500/ws?clientid=${loggedInUserId}&members=${unique_members}`);
    socketRef.current.onmessage = function (event) {
      try {
        const incomingMessage = JSON.parse(event.data);
        setMessages((prev) => [...prev, incomingMessage]);
      } catch (err) {
        console.error("WebSocket parse error:", err);
      }
    };
    };
    prepare()
  }, [loggedInUserId, userId, token]);

  const handleInputChange = (e) => {
    setMessageText(e.target.value);
  };

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
  };

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
        setSearchTerm("");
      } catch (error) {
        console.error("Error fetching search results:", error);
      }
    }
  };

  const handleUserSelect = (user) => {
    setSelectedUsers([...selectedUsers, user]);
    setShowResultsModal(false);
  };

  const handleAddMemebrsToGroup = async () => {
    const token = localStorage.getItem("jwtToken");
    const payload = {
      name: username,
      members: selectedUsers.map((user) => user.userid),
    };

    try {
      const response = await axios.put(
        `http://localhost:8080/v1/group/addmembers/${userId}`,
        payload,
        {
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
        }
      );

      if (response.status === 200 || response.status === 201) {
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

  const handleShowAddMemberModalClick = () => {
    setShowGroupModal(true);
  };

  const handleGroupModalClose = () => {
    setShowGroupModal(false);
  };

  const scrollToBottom = () => {
    if (chatMessagesRef.current) {
      chatMessagesRef.current.scrollTop = chatMessagesRef.current.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async () => {
    if (!messageText.trim()) return;

    const tempId = `temp-${Date.now()}`;
    const newMessage = {
      id: tempId,
      sender_id: Number(loggedInUserId),
      receiver_id: Number(userId),
      text: messageText.trim(),
      date: new Date().toISOString(),
      sender_name: "You",
      status: "sending",
    };

    setMessages((prevMessages) => [...prevMessages, newMessage]);
    setMessageText("");

    try {
      const response = await axios.post(
        `http://localhost:8080/v1/chat/group/${userId}`,
        {
          sender_id: Number(loggedInUserId),
          receiver_id: Number(userId),
          text: messageText.trim(),
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        const websocketPayload = {
          sender_id: newMessage.sender_id,
          receiver_id: newMessage.receiver_id,
          text: newMessage.text,
          date: newMessage.date,
          sender_name: response.data.data
        };
        socketRef.current.send(JSON.stringify(websocketPayload));
      }      

      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: "sent" } : msg
        )
      );
    } catch (error) {
      console.error("Error sending message:", error);
      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: "failed" } : msg
        )
      );
    }
  };

  const handleGroupNameClick = async () => {
    try {
      const response = await axios.get(
        `http://localhost:8080/v1/group/info/${userId}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );
      setGroupInfo(response.data.data);
      setShowGroupInfo(true);
      return response.data.data;
    } catch (error) {
      console.error("Error fetching group info:", error);
    }
  };

  const ChatMessage = ({ message, isOwnMessage }) => {
    return (
      <div className={`chat-message ${isOwnMessage ? "own" : "other"}`}>
        <div className="sender-info">
          <span
            className="sender-name"
            style={{ cursor: "pointer", color: "#0d6efd", display: "block" }}
            onClick={() => handleUserClick(message.sender_id)}
          >
            {message.sender_name}
          </span>
          <span className="message-date">
            {new Date(message.date).toLocaleString()}
          </span>
        </div>
        {message.text && (
          <p className={`message ${isOwnMessage ? "own" : "other"}`}>
            {message.text}
            {message.status === "sending" && (
              <span className="status"> (Sending...)</span>
            )}
            {message.status === "failed" && (
              <span className="status"> (Failed to send)</span>
            )}
          </p>
        )}
        {message.post && <PostCard post={message.post} />}
      </div>
    );
  };

  return (
    <div className="chat-container">
      <header
        className="chat-header"
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <h2
          style={{ cursor: "pointer", color: "#1e88e5" }}
          onClick={handleGroupNameClick}
        >
          {username}
        </h2>
        <h4
          onClick={handleShowAddMemberModalClick}
          style={{ cursor: "pointer", marginTop: "6px" }}
        >
          👥+
        </h4>
      </header>

      <div className="chat-messages" ref={chatMessagesRef}>
        {messages &&
          messages.map((msg, index) => (
            <ChatMessage
              key={index}
              message={msg}
              isOwnMessage={msg.sender_id === loggedInUserId}
            />
          ))}
      </div>

      <div className="chat-input">
        <input
          type="text"
          className="message-input"
          value={messageText}
          onChange={handleInputChange}
          placeholder="Type your message..."
        />
        <button className="send-button" onClick={handleSendMessage}>
          Send
        </button>
      </div>

      {showGroupInfo && (
        <GroupInfoModal
          groupInfo={groupInfo}
          onClose={() => setShowGroupInfo(false)}
        />
      )}

      <LikedUsersModal
        show={showResultsModal}
        likedUsers={searchResults}
        onClose={() => setShowResultsModal(false)}
        title="Search Results"
        emptytitle="users"
        {...(true && { onUserSelect: handleUserSelect })}
      />

      {showGroupModal && (
        <div className="modal fade show d-block" tabIndex="-1" role="dialog">
          <div className="modal-dialog" role="document">
            <div className="modal-content">
              <div className="modal-header">
                <h5 className="modal-title">Add Members</h5>
                <button
                  type="button"
                  className="btn-close"
                  onClick={handleGroupModalClose}
                  aria-label="Close"
                ></button>
              </div>
              <div className="modal-body">
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
                  onClick={handleAddMemebrsToGroup}
                >
                  Add Members
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupChatPage;
