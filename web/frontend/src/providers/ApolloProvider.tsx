import React from 'react'
import { ApolloProvider } from '@apollo/client'
import { apolloClient } from '@/lib/apollo-client'

interface ApolloProviderProps {
  children: React.ReactNode
}

const CustomApolloProvider: React.FC<ApolloProviderProps> = ({ children }) => {
  return <ApolloProvider client={apolloClient}>{children}</ApolloProvider>
}

export default CustomApolloProvider