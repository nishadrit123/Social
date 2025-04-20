import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import axios from "axios";
import "../style/ChatPage.css";
import PostCard from "../components/PostCard";

const ChatPage = () => {
  const { userId, username } = useParams();
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const token = localStorage.getItem("jwtToken");
  const decoded = jwtDecode(token);
  const loggedInUserId = decoded.sub;

  useEffect(() => {
    const fetchChat = async () => {
      try {
        const response = await axios.get(
          `http://localhost:8080/v1/chat/user/${userId}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
        setMessages(response.data.data);
      } catch (error) {
        console.error("Error fetching chat:", error);
      }
    };

    fetchChat();
  }, [userId, token]);

  const ChatMessage = ({ message, isOwnMessage }) => {
    return (
      <div className={`chat-message ${isOwnMessage ? "own" : "other"}`}>
        <div className="message-date">
          {new Date(message.date).toLocaleString()}
        </div>
        {message.text && (
          <p className={`message ${isOwnMessage ? "own" : "other"}`}>
            {message.text}
          </p>
        )}
        {message.post && <PostCard post={message.post} />}
      </div>
    );
  };

  const handleSendMessage = () => {
    // Implement send message functionality here
    // This will be discussed later as per your guidance
  };

  return (
    <div className="chat-container">
      <header className="chat-header">
        <h2>{username}</h2>
      </header>
      <div className="chat-messages">
        {messages && messages.map((msg, index) => (
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
          placeholder="Type your message..."
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
        />
        <button className="send-button" onClick={handleSendMessage}>
          Send
        </button>
      </div>
    </div>
  );
};

export default ChatPage;
