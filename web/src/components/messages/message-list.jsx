import React, { Component } from 'react';
import PropTypes from 'prop-types';

import Message from './message';

class MessageList extends Component {
  render() {
    return (
      <ul>
        {this.props.messages.map(message => {
          return <Message key={message.id} message={message} />;
        })}
      </ul>
    );
  }
}

// PropTypes
MessageList.propTypes = {
  messages: PropTypes.array.isRequired
};

export default MessageList;
