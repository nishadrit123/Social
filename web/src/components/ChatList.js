import React, { useEffect, useState } from "react";
import axios from "axios";
import "bootstrap/dist/css/bootstrap.min.css";

const ChatList = ({ onChatSelect }) => {
  const [chatList, setChatList] = useState([]);

  useEffect(() => {
    const fetchChatList = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const response = await axios.get(
          "http://localhost:8080/v1/chat/chatwindow",
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );
        setChatList(response.data.data || []);
      } catch (error) {
        console.error("Error fetching chat list:", error);
      }
    };

    fetchChatList();
  }, []);

  return (
    <div className="card overflow-auto">
      <div className="card-header">
        <h5 className="mb-0">Chats & Groups</h5>
      </div>
      <ul className="list-group list-group-flush">
        {chatList.map((chat) => (
          <li
            key={chat.userid}
            className="list-group-item list-group-item-action"
            onClick={() => onChatSelect(chat)}
            style={{ cursor: "pointer" }}
          >
            {chat.is_group ? (
              <>👥&nbsp;&nbsp;&nbsp;{chat.username}</>
            ) : (
              <>👤&nbsp;&nbsp;&nbsp;{chat.username}</>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ChatList;
