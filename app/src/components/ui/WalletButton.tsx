import { useEffect, useState } from "react";
import { useWallet, useConnection } from "@solana/wallet-adapter-react";
import { useWalletModal } from "@solana/wallet-adapter-react-ui";
import { LAMPORTS_PER_SOL } from "@solana/web3.js";

export default function WalletButton() {
  const { connected, publicKey, disconnect, connecting } = useWallet();
  const { connection } = useConnection();
  const { setVisible } = useWalletModal();
  const [balance, setBalance] = useState<number | null>(null);

  useEffect(() => {
    if (!connected || !publicKey) {
      setBalance(null);
      return;
    }
    connection.getBalance(publicKey).then((lamports) => {
      setBalance(lamports / LAMPORTS_PER_SOL);
    });
  }, [connected, publicKey, connection]);

  if (connecting) {
    return (
      <button
        disabled
        className="h-9 px-4 rounded-xl text-sm font-medium bg-green-700/50 text-green-300 cursor-not-allowed"
      >
        Connecting…
      </button>
    );
  }

  if (connected && publicKey) {
    const short = `${publicKey.toBase58().slice(0, 4)}…${publicKey.toBase58().slice(-4)}`;
    return (
      <button
        onClick={() => disconnect()}
        className="h-9 px-4 rounded-xl text-sm font-medium bg-green-600 hover:bg-green-700 text-white transition-colors flex items-center gap-2"
      >
        {balance !== null && (
          <span className="text-green-200 text-xs font-normal">
            {balance.toFixed(2)} SOL
          </span>
        )}
        <span>{short}</span>
      </button>
    );
  }

  return (
    <button
      onClick={() => setVisible(true)}
      className="h-9 px-4 rounded-xl text-sm font-medium bg-green-600 hover:bg-green-700 text-white transition-colors"
    >
      Connect Wallet
    </button>
  );
}
