/*
 * PFAC kernel — Parallel Failureless Aho-Corasick.
 *
 * Każdy work-item (wątek GPU) startuje od innego offsetu w tekście
 * i samodzielnie przechodzi po drzewie trie. Brak failure links oznacza,
 * że przy pierwszym nierozpoznanym znaku wątek natychmiast się zatrzymuje.
 * Nie ma żadnej synchronizacji między wątkami — klasyczna równoległość danych.
 *
 * Argumenty kernela:
 *   text         - bufor z tekstem (bajty)
 *   text_len     - długość tekstu
 *   goto_table   - spłaszczona tablica przejść [numStates * alphabetSize], -1 = martwy
 *   has_output   - [numStates], 1 jeśli stan jest stanem akceptującym
 *   alphabet_size- rozmiar alfabetu (256)
 *   total_matches- globalny licznik dopasowań (atomic)
 */
__kernel void pfac(
    __global const uchar  *text,
    const int              text_len,
    __global const int    *goto_table,
    __global const int    *has_output,
    const int              alphabet_size,
    __global volatile int *total_matches
) {
    int gid = get_global_id(0);
    if (gid >= text_len) return;

    int state = 0;
    for (int pos = gid; pos < text_len; pos++) {
        int c = (int)text[pos];
        int next = goto_table[state * alphabet_size + c];
        if (next < 0) break;   /* martwy stan — żaden wzorzec nie startuje z tej pozycji */
        state = next;
        if (has_output[state]) {
            atomic_inc(total_matches);
        }
    }
}
