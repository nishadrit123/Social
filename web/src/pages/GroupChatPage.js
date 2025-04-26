import React, { useEffect, useState, useRef } from "react";
import { useParams } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import axios from "axios";
import "../style/ChatPage.css";
import PostCard from "../components/PostCard";
import GroupInfoModal from "../components/GroupInfoModal";
import { useNavigate } from "react-router-dom";

const GroupChatPage = () => {
  const { userId, username } = useParams();
  const [messages, setMessages] = useState([]);
  const [messageText, setMessageText] = useState("");
  const [groupInfo, setGroupInfo] = useState(null);
  const [showGroupInfo, setShowGroupInfo] = useState(false);
  const chatMessagesRef = useRef(null);

  const token = localStorage.getItem("jwtToken");
  const decoded = jwtDecode(token);
  const loggedInUserId = decoded.sub;
  const navigate = useNavigate();

  useEffect(() => {
    const fetchGroupChat = async () => {
      try {
        const response = await axios.get(`http://localhost:8080/v1/chat/group/${userId}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        setMessages(response.data.data);
      } catch (error) {
        console.error("Error fetching group chat:", error);
      }
    };

    fetchGroupChat();
  }, [userId, token]);

  const handleInputChange = (e) => {
    setMessageText(e.target.value);
  };

  const handleUserClick = (userid) => {
    navigate(`/profile/${userid}`);
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
      status: 'sending',
    };

    setMessages((prevMessages) => [...prevMessages, newMessage]);
    setMessageText('');

    try {
      await axios.post(
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

      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: 'sent' } : msg
        )
      );
    } catch (error) {
      console.error('Error sending message:', error);
      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: 'failed' } : msg
        )
      );
    }
  };

  const handleGroupNameClick = async () => {
    try {
      const response = await axios.get(`http://localhost:8080/v1/group/info/${userId}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      setGroupInfo(response.data.data);
      setShowGroupInfo(true);
    } catch (error) {
      console.error("Error fetching group info:", error);
    }
  };

  const ChatMessage = ({ message, isOwnMessage }) => {
    return (
      <div className={`chat-message ${isOwnMessage ? 'own' : 'other'}`}>
        <div className="sender-info">
          <span className="sender-name" style={{ cursor: "pointer", color: "#0d6efd", display: "block" }}
          onClick={() => handleUserClick(message.sender_id)}>
            {message.sender_name}
          </span>
          <span className="message-date">
            {new Date(message.date).toLocaleString()}
          </span>
        </div>
        {message.text && (
          <p className={`message ${isOwnMessage ? 'own' : 'other'}`}>
            {message.text}
            {message.status === 'sending' && <span className="status"> (Sending...)</span>}
            {message.status === 'failed' && <span className="status"> (Failed to send)</span>}
          </p>
        )}
        {message.post && <PostCard post={message.post} />}
      </div>
    );
  };

  return (
    <div className="chat-container">
      <header className="chat-header">
        <h2 style={{ cursor: "pointer", color: "#1e88e5" }} onClick={handleGroupNameClick}>
          {username}
        </h2>
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
    </div>
  );
};

export default GroupChatPage;
