import React, { useEffect, useState, useRef } from "react";
import { useParams } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import axios from "axios";
import "../style/ChatPage.css";
import PostCard from "../components/PostCard";

const ChatPage = () => {
  const { userId, username } = useParams();
  const [messages, setMessages] = useState([]);
  const [messageText, setMessageText] = useState("");
  const chatMessagesRef = useRef(null);
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

  const handleInputChange = (e) => {
    setMessageText(e.target.value);
  };

  const scrollToBottom = () => {
    if (chatMessagesRef.current) {
      chatMessagesRef.current.scrollTop = chatMessagesRef.current.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);  

  // const handleSendMessage = async () => {
  //   if (!messageText.trim()) return;

  //   const payload = {
  //     sender_id: Number(loggedInUserId), 
  //     receiver_id: Number(userId), 
  //     text: messageText.trim(),
  //   };

  //   try {
  //     const response = await axios.post(
  //       `http://localhost:8080/v1/chat/user/${userId}`,
  //       payload,
  //       {
  //         headers: {
  //           Authorization: `Bearer ${token}`,
  //         },
  //       }
  //     );
  //     setMessageText(""); // Clear the input field
  //   } catch (error) {
  //     console.error("Error sending message:", error);
  //   }
  // };

  const handleSendMessage = async () => {
    if (!messageText.trim()) return;
  
    const tempId = `temp-${Date.now()}`;
    const newMessage = {
      id: tempId,
      sender_id: Number(loggedInUserId),
      receiver_id: Number(userId),
      text: messageText.trim(),
      date: new Date().toISOString(),
      status: 'sending',
    };
  
    // Add the temporary message to the UI
    setMessages((prevMessages) => [...prevMessages, newMessage]);
    setMessageText('');
  
    try {
      await axios.post(
        `http://localhost:8080/v1/chat/user/${userId}`,
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
  
      // Since the API doesn't return the saved message, we assume success
      // Optionally, you can update the status to 'sent' if you wish
      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: 'sent' } : msg
        )
      );
    } catch (error) {
      console.error('Error sending message:', error);
      // Update the message status to 'failed'
      setMessages((prevMessages) =>
        prevMessages.map((msg) =>
          msg.id === tempId ? { ...msg, status: 'failed' } : msg
        )
      );
    }
  };  

  // const ChatMessage = ({ message, isOwnMessage }) => {
  //   return (
  //     <div className={`chat-message ${isOwnMessage ? "own" : "other"}`}>
  //       <div className="message-date">
  //         {new Date(message.date).toLocaleString()}
  //       </div>
  //       {message.text && (
  //         <p className={`message ${isOwnMessage ? "own" : "other"}`}>
  //           {message.text}
  //         </p>
  //       )}
  //       {message.post && <PostCard post={message.post} />}
  //     </div>
  //   );
  // };

  const ChatMessage = ({ message, isOwnMessage }) => {
    return (
      <div className={`chat-message ${isOwnMessage ? 'own' : 'other'}`}>
        <div className="message-date">
          {new Date(message.date).toLocaleString()}
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
        <h2>{username}</h2>
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
    </div>
  );
};

export default ChatPage;
