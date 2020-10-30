import React, { Component } from 'react';

import ChannelSection from './components/channels/channel-section';
import UserSection from './components/users/user-section';
import MessageSection from './components/messages/message-section';
import Socket from './socket';

import './App.css';

class App extends Component {
  constructor() {
    super();
    this.state = {
      channels: [],
      users: [],
      messages: [],
      activeChannel: {},
      connected: false,
    };
  }

  componentDidMount() {
    let ws = new WebSocket(process.env.REACT_APP_SERVER_URL);
    let socket = (this.socket = new Socket(ws));
    socket.on('connect', this.onConnect);
    socket.on('disconnect', this.onDisconnect);
    socket.on('channel add', this.onAddChannel);
    socket.on('user add', this.onAddUser);
    socket.on('user edit', this.onEditUser);
    socket.on('user remove', this.onRemoveUser);
    socket.on('message add', this.onMessageAdd);
    socket.on('error', this.onError);
  }

  onError = (err) => {
    console.error('whoops! there was an error');
  };

  onMessageAdd = (message) => {
    let { messages } = this.state;
    messages.push(message);
    this.setState({ messages });
  };

  onAddUser = (user) => {
    let { users } = this.state;
    users.push(user);
    this.setState({ users });
  };

  onEditUser = (editUser) => {
    let { users } = this.state;
    users = users.map((user) => {
      if (editUser.id === user.id) {
        return editUser;
      }
      return user;
    });
    this.setState({ users });
  };

  onRemoveUser = (removeUser) => {
    let { users } = this.state;
    users = users.filter((user) => {
      return user.id !== removeUser.id;
    });
    this.setState({ users });
  };

  onConnect = () => {
    this.setState({ connected: true });
    this.socket.emit('channel subscribe');
    this.socket.emit('user subscribe');
  };

  onDisconnect = () => {
    this.setState({ connected: true });
  };

  onAddChannel = (channel) => {
    let { channels } = this.state;
    channels.push(channel);
    this.setState({ channels });
  };

  addChannel = (name) => {
    this.socket.emit('channel add', { name });
  };

  setChannel = (activeChannel) => {
    this.setState({ activeChannel });
    this.socket.emit('message unsubscribe');
    this.setState({ messages: [] });
    this.socket.emit('message subscribe', { channelId: activeChannel.id });
  };

  setUserName = (name) => {
    this.socket.emit('user edit', { name });
  };

  addMessage = (body) => {
    let { activeChannel } = this.state;
    this.socket.emit('message add', {
      channelId: activeChannel.id,
      body,
    });
  };

  render() {
    return (
      <div className="app">
        <div className="nav">
          <ChannelSection
            {...this.state}
            addChannel={this.addChannel}
            setChannel={this.setChannel}
          />
          <UserSection {...this.state} setUserName={this.setUserName} />
        </div>
        <MessageSection {...this.state} addMessage={this.addMessage} />
      </div>
    );
  }
}

export default App;
