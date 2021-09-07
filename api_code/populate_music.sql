INSERT INTO music("songID", "songName", "artist", "length")
VALUES (2001, 'Promiscuous', 'Nelly Furtado', '00:04:02'),
        (2002, 'Just A Lil Bit', '50 Cent', '00:03:57'),
        (2003, 'Gangstas Paradise', 'Coolio, L.V.', '00:04:00'),
        (2004, 'La Bamba', 'Los Lobos', '00:02:54'),
        (2005, 'Sex Bomb', 'Pablo Lopez', '00:03:35'),
        (2006, 'Pump It', 'Black Eyed Peas', '00:03:33'),
        (2007, 'Gimme! Gimme! Gimme! (A Man After Midnight)', 'ABBA', '00:04:02');

INSERT INTO listens("userID", "songID")
VALUES (1002, 2004), 
        (1004, 2001), 
        (1005, 2006), 
        (1001, 2005);